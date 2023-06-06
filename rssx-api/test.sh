#!/bin/sh
echo "socat starting..."
socat TCP-LISTEN:3389,fork TCP:192.168.55.2:3389