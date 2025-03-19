# Token Ring Algorithm Test

## Overview
This project is an implementation of the **Token Ring Algorithm** in Golang, to verify feasability for usage in multiple gateway support for CVMFS. The algorithm ensures mutual exclusion in a distributed system by passing a token among nodes in a logical ring structure. This test setup allows nodes to pass the token around, and communicate when a node fails.

## Features
- Implements a token-passing mechanism access to a common resource (`common.txt`).
- Uses HTTP for inter-node communication.
- Handles node failures by removing inactive nodes from the ring.
- Implements an acknowledgment system to confirm token receipt. If no acknowledgement is given within a timer, the token is instead passed to the next node. This prevents a deadlock when a failing node is given the token.

## Usage
### Setup
1. Run the `resetNodes.sh` script. This prepares the three directories in which the nodes can be run, and the required `ring_ports.txt` files which contain the list of ports on which nodes are running.

### Running the nodes
**Do the first two steps for each node!**
1. Go into the node running directory
   ```sh
   cd <<nodeNumber>>
   ```
2. Start the program:
   ```sh
   PORT=${PWD##*/} go run ../src
   ```
   This sets the port to the name of the current directory
3. Hand the token to the first node
   ```sh
   curl -X POST http://localhost:5000/token
   ```