#!/bin/bash

# Reset the ring ports files
RING_PORTS_FILE="ring_ports.txt"

for i in {5000..5002}; do
  mkdir -p $i
  for j in {1..2}; do
    sudo rm -f $i/repo$j-$RING_PORTS_FILE
    touch $i/repo$j-$RING_PORTS_FILE
    echo "5000
5001
5002" >> $i/repo$j-$RING_PORTS_FILE
  done
done

# Reset the output dir
sudo rm -rf output