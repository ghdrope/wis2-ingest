#!/usr/bin/env python3
"""
WIS2 MQTT Ingest Client

This script subscribes to MQTT topics, downloads incoming WIS2 data files
from URLs provided in messages, and organizes them in a structured output
directory by capture day and hour (for known files) or into 'unknown' for .bin files.
It supports configuration via command-line, environment variables, or an INI config file.

Usage:
  wis2_ingest.py [--config=<file>] [--host=<host>] [--port=<port>] [--topic=<topic>...]
                 [--username=<username>] [--password=<password>]
                 [--inpath=<inpath>] [--outpath=<outpath>] [--logfile=<logfile>]
  wis2_ingest.py (-h | --help)

Options:
  -h --help              Show this help message.
  --config=<file>        Config file [default: ./mqtt_config.ini]
  --host=<host>          MQTT server host
  --port=<port>          MQTT server port
  --topic=<topic>        MQTT topic (repeatable)
  --username=<username>  MQTT username
  --password=<password>  MQTT password
  --inpath=<inpath>      Incoming files directory
  --outpath=<outpath>    Output files directory
  --logfile=<logfile>    Log file base path
"""

# ==========================================================
# Imports
# ==========================================================
import os
import sys
import ssl
import json
import shutil
from datetime import datetime
from threading import Lock
import configparser
from urllib.request import urlretrieve
from urllib.parse import urlparse
from urllib.error import URLError

import paho.mqtt.client as mqtt
from docopt import docopt

# ==========================================================
# Globals
# ==========================================================
data_ids = []          # Track processed data_ids to avoid duplicates
incoming_path = ""     # Temporary storage path for downloaded files
output_path = ""       # Base directory where files are organized (WIS2_arrived)
logger = None          # Logger instance
log_file_base = None   # Base path for log files
connection_info = {}   # Stores resolved configuration for logging

# ==========================================================
# Daily Rotating Log Writer
# ==========================================================
class DailyRotatingWriter:
    """
    Thread-safe logger that rotates daily based on UTC date.
    
    Each day, a new file is created with the format:
        YYYYMMDD_<basename>.log
    """

    def __init__(self, base_path):
        self.base_path = base_path
        self.current_day = None
        self.file = None
        self.lock = Lock()
        self._open_new_file()

    def _get_filename(self, day):
        """Construct full log filename for the given date."""
        directory = os.path.dirname(self.base_path) or "."
        basename = os.path.basename(self.base_path)
        prefix = day.strftime("%Y%m%d")
        return os.path.join(directory, f"{prefix}_{basename}")

    def _open_new_file(self):
        """Open a new log file for the current UTC day."""
        today = datetime.utcnow().date()
        filename = self._get_filename(datetime.utcnow())
        os.makedirs(os.path.dirname(filename), exist_ok=True)
        self.file = open(filename, "a", encoding="utf-8")
        self.current_day = today

    def write(self, level, message):
        """Thread-safe write to log, rotating daily if needed."""
        with self.lock:
            today = datetime.utcnow().date()
            if today != self.current_day or not os.path.exists(self.file.name):
                try:
                    self.file.close()
                except Exception:
                    pass
                self._open_new_file()
            timestamp = datetime.utcnow().strftime("%Y%m%d:%H:%M:%S")
            self.file.write(f"{timestamp} {level}: {message}\n")
            self.file.flush()

    def close(self):
        if self.file:
            self.file.close()

# ==========================================================
# Logger
# ==========================================================
class Logger:
    """Simple logger interface: info, warning, error levels."""
    def __init__(self, logfile):
        self.writer = DailyRotatingWriter(logfile)

    def info(self, msg):
        self.writer.write("INFO", msg)
    
    def warning(self, msg):
        self.writer.write("WARNING", msg)
    
    def error(self, msg):
        self.writer.write("ERROR", msg)

# ==========================================================
# Connection header logging
# ==========================================================
def write_connection_header():
    """Log resolved MQTT configuration and connection info."""
    if not logger or not connection_info:
        return
    logger.info(
        f"Resolved MQTT configuration: host={connection_info['host']}, "
        f"port={connection_info['port']}, topic={connection_info['topic']}, "
        f"username={connection_info['username']}, "
        f"inpath={connection_info['inpath']}, "
        f"outpath={connection_info['outpath']}, "
        f"logfile={connection_info['logfile']}"
    )
    logger.info(f"Connecting to {connection_info['host']}:{connection_info['port']}")

def setup_logger():
    """Initialize the global logger instance."""
    global logger
    logger = Logger(log_file_base)

# ==========================================================
# MQTT Callbacks
# ==========================================================
def on_connect(client, userdata, flags, rc):
    """Called when the MQTT client connects to the broker."""
    logger.info(f"Connected with result code {rc}")
    if rc == 0:
        topics = [(t, 1) for t in userdata["topic"]]
        client.subscribe(topics)
        logger.info(f"Subscribed to topics: {userdata['topic']}")
    else:
        logger.error("Failed to connect to MQTT broker")

def on_message(client, userdata, msg):
    """
    Called when a message is received.

    Downloads files from 'canonical' links in the message payload and
    organizes them:
        - Known files (not .bin) → <DAY>/<HOUR> (UTC)
        - .bin files → unknown/
    """
    global data_ids

    # Log raw payload for debugging
    try:
        payload_str = msg.payload.decode("utf-8", errors="ignore")
    except Exception:
        payload_str = str(msg.payload)
    logger.info(f"RAW message received on topic {msg.topic}: {payload_str}")

    try:
        payload = json.loads(payload_str)
    except Exception as e:
        logger.warning(f"Invalid JSON payload: {e}")
        return

    # Use data_id to avoid duplicate processing
    props = payload.get("properties", {})
    data_id = props.get("data_id")
    if not data_id:
        logger.warning("Message without data_id")
        return
    if data_id in data_ids:
        return
    data_ids.append(data_id)
    if len(data_ids) > 50: # Keep only last 50 to limit memory usage
        data_ids.pop(0)

    # Capture time in UTC
    capture_time = datetime.utcnow()
    day_str = capture_time.strftime("%d%m%Y")
    hour_str = capture_time.strftime("%H")

    # Process each link in message
    for link in payload.get("links", []):
        if link.get("rel") != "canonical":
            continue
        url = link.get("href")
        if not url:
            continue

        filename = os.path.basename(urlparse(url).path)
        in_file = os.path.join(incoming_path, filename)

        # Download the file
        try:
            urlretrieve(url, in_file)
        except URLError:
            logger.warning(f"Download failed: {url}")
            continue

        # Determine target directory
        if filename.lower().endswith(".bin"):
            out_dir = os.path.join(output_path, "unknown")
        else:
            out_dir = os.path.join(output_path, day_str, hour_str)

        # Ensure directory exists
        os.makedirs(out_dir, exist_ok=True)

        # Move file to final destination
        shutil.move(in_file, os.path.join(out_dir, filename))
        logger.info(f"File stored: {filename} in {out_dir}")

# ==========================================================
# Utilities
# ==========================================================
def sanitize_topic(topic):
    """Convert MQTT topic into a safe directory name."""
    return topic.replace("/", "_").replace("+", "plus").replace("#", "hash")

def load_config(path):
    """Load configuration from INI file if it exists."""
    if not os.path.exists(path):
        return {}
    config = configparser.ConfigParser()
    config.read(path)
    return config["MQTT"] if "MQTT" in config else {}

# ==========================================================
# Argument Parsing
# ==========================================================
def parse_args():
    """Parse command-line arguments, environment variables, and config file."""
    args = docopt(__doc__)
    config = load_config(args["--config"])

    def resolve(value, env, key):
        """Resolve parameter: CLI > ENV > Config"""
        return value or os.getenv(env) or config.get(key)

    host = resolve(args["--host"], "REMOTE_MQTT_HOST", "host")
    port = resolve(args["--port"], "REMOTE_MQTT_PORT", "port")
    topic = args["--topic"] or os.getenv("REMOTE_MQTT_TOPIC") or config.get("topic")
    username = resolve(args["--username"], "REMOTE_MQTT_USERNAME", "username")
    password = resolve(args["--password"], "REMOTE_MQTT_PASSWORD", "password")
    inpath = resolve(args["--inpath"], "INCOMING_DIRECTORY", "inpath")
    outpath = resolve(args["--outpath"], "OUTPUT_DIRECTORY", "outpath")
    logfile = resolve(args["--logfile"], "LOG_FILE", "logfile")

    # Support multiple topics separated by ';'
    if isinstance(topic, str):
        topic = topic.split(";")

    if not host or not port:
        print("ERROR: MQTT host and port are required")
        return None

    return host, port, topic, username, password, inpath, outpath, logfile

# ==========================================================
# Main entry point
# ==========================================================
def main():
    global incoming_path, output_path, log_file_base, connection_info

    params = parse_args()
    if not params:
        sys.exit(1)

    host, port, topic, username, password, incoming_path, output_path, log_file_base = params
    port = int(port)

    # Ensure directories exist
    os.makedirs(incoming_path, exist_ok=True)
    os.makedirs(output_path, exist_ok=True)

    # Store configuration for logging
    connection_info = {
        "host": host,
        "port": port,
        "topic": topic,
        "username": username,
        "inpath": incoming_path,
        "outpath": output_path,
        "logfile": log_file_base,
    }

    # Initialize logger and write header
    setup_logger()
    write_connection_header()
    logger.info("==== WIS2 MQTT INGEST START ====")

    # Setup MQTT client
    client = mqtt.Client(userdata={"topic": topic})
    client.tls_set(tls_version=ssl.PROTOCOL_TLSv1_2)
    if username and password:
        client.username_pw_set(username, password)

    client.on_connect = on_connect
    client.on_message = on_message

    # Connect to broker and loop forever
    client.connect(host, port, 60)
    client.loop_forever()

if __name__ == "__main__":
    main()
    