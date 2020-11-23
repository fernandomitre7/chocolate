#!/bin/sh
SERVICE_NAME="chocolate"
if [ -f bin/PID ]; then 
        PID=`cat bin/PID`
        echo "Got existing PID ${PID} from File"
else 
    EGREP=`ps -a | egrep ${SERVICE_NAME}`
    if [[ "${EGREP}" != *"egrep"* ]]; then
        PID=`ps -a | egrep ${SERVICE_NAME} | awk '{print $1}'`
        echo "Got existing PID ${PID} from process"
    fi
fi

echo "Stopping ${SERVICE_NAME} pid ${PID}..."
kill ${PID}
echo "Stopped... Moving log files"
TODAY=`date +%s%N | cut -b1-13`
if [ -f logs/${SERVICE_NAME}.log ]; then 
    mv logs/${SERVICE_NAME}.log logs/${SERVICE_NAME}.${TODAY}.log; 
fi

if [ -f bin/PID ]; then 
    # for security if file still exists then just truncate it
    cat /dev/null > bin/PID
fi
