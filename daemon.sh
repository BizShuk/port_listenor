#!/bin/bash
#
# daemon.sh — Run port_listenor monitor as a daemon
#
# Usage:
#   ./daemon.sh start   Start the daemon
#   ./daemon.sh stop    Stop the daemon
#   ./daemon.sh restart Restart the daemon
#   ./daemon.sh status  Show daemon status
#
# Environment:
#   MIMIR_URL       Override mimir_url in config (e.g., http://localhost:9009/otlp/v1/metrics)
#   MIMIR_PORT      Override metrics_port (default: from config)
#

NAME="port_listenor"
BINARY="./port_listenor"
PIDFILE="/tmp/${NAME}.pid"
LOGFILE="/tmp/${NAME}.log"
CONFIG_PATH="${HOME}/.config/port_listenor/settings.json"

# Load MIMIR_URL from config if not set
load_config() {
    if [ -f "${CONFIG_PATH}" ] && [ -z "${MIMIR_URL}" ]; then
        MIMIR_URL=$(grep -o '"mimir_url"[[:space:]]*:[[:space:]]*"[^"]*"' "${CONFIG_PATH}" 2>/dev/null | sed 's/.*"mimir_url"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    fi
}

do_start() {
    load_config

    if [ -f "${PIDFILE}" ]; then
        PID=$(cat "${PIDFILE}")
        if kill -0 "${PID}" 2>/dev/null; then
            echo "${NAME} is already running (PID: ${PID})"
            return 1
        fi
        rm -f "${PIDFILE}"
    fi

    echo "Starting ${NAME}..."

    ARGS="monitor"
    if [ -n "${MIMIR_URL}" ]; then
        # Note: mimir_url must be set in config file, not via CLI flag currently
        echo "MIMIR_URL=${MIMIR_URL} (set in ${CONFIG_PATH})"
    fi

    # Note: The current CLI does not support --mimir-endpoint flag directly
    # Set mimir_endpoint in ~/.config/port_listenor/settings.json instead
    nohup ${BINARY} ${ARGS} >> "${LOGFILE}" 2>&1 &
    PID=$!
    echo "${PID}" > "${PIDFILE}"
    echo "${NAME} started (PID: ${PID})"
    echo "Log: ${LOGFILE}"
}

do_stop() {
    if [ ! -f "${PIDFILE}" ]; then
        echo "${NAME} is not running (no PID file)"
        return 1
    fi

    PID=$(cat "${PIDFILE}")
    if kill -0 "${PID}" 2>/dev/null; then
        echo "Stopping ${NAME} (PID: ${PID})..."
        kill "${PID}"
        sleep 1
        if kill -0 "${PID}" 2>/dev/null; then
            echo "Force killing ${NAME}..."
            kill -9 "${PID}"
        fi
        rm -f "${PIDFILE}"
        echo "${NAME} stopped"
    else
        echo "${NAME} is not running (stale PID file)"
        rm -f "${PIDFILE}"
    fi
}

do_status() {
    if [ -f "${PIDFILE}" ]; then
        PID=$(cat "${PIDFILE}")
        if kill -0 "${PID}" 2>/dev/null; then
            echo "${NAME} is running (PID: ${PID})"
            return 0
        fi
        echo "${NAME} is not running (stale PID: ${PID})"
        return 1
    fi
    echo "${NAME} is not running"
    return 1
}

case "${1}" in
    start)
        do_start
        ;;
    stop)
        do_stop
        ;;
    restart)
        do_stop
        sleep 1
        do_start
        ;;
    status)
        do_status
        ;;
    *)
        echo "Usage: ${0} {start|stop|restart|status}"
        exit 1
        ;;
esac
