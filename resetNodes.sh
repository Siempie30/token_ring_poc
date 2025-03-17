#!/bin/bash

# Reset the ring ports file
RING_PORTS_FILE="ring_ports.txt"
rm -f $RING_PORTS_FILE
touch $RING_PORTS_FILE
cat <<EOT >> $RING_PORTS_FILE
5000
5001
5002
EOT