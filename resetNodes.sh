#!/bin/bash

# Reset the ring ports files
RING_PORTS_FILE="ring_ports.txt"
rm -rf ./500*

for i in {5000..5002}; do
  mkdir -p $i
  touch $i/$RING_PORTS_FILE
  echo "5000
5001
5002" >> $i/$RING_PORTS_FILE
done