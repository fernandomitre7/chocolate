#!/bin/sh
# TODO: move this to function
SERVICE_NAME="chocolate"
EGREP=`ps -a | egrep ${SERVICE_NAME}`

if [[ "${EGREP}" != *"egrep"* ]]; then
    PID=`ps -a | egrep ${SERVICE_NAME} | awk '{print $1}'`
    echo "Got existing PID ${PID} from process"
elif [ -f bin/PID ]; then 
    PID=`cat bin/PID`
    echo "Got existing PID ${PID} from File" 
fi

if [ ! -z "${PID}" ]; then 
    echo "API already running"
    exit 1
fi

if [ ! -d "logs" ]; then
    echo "creating logs directory"
    mkdir logs
fi

if [ -f bin/${SERVICE_NAME} ]; then
    nohup bin/${SERVICE_NAME} >> logs/${SERVICE_NAME}.out 2>&1 &
else
    echo "API binary not found, try running make"
fi

echo "Running"