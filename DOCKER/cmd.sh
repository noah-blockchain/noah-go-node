#!/bin/bash

cleanup ()
{
    OUTPUT=`"/blockchain/stop.sh"`
    echo $OUTPUT
}

trap cleanup SIGINT SIGTERM

while true; do sleep 1 ; echo ""; done