# Token Ring Algorithm Test

## Overview
This project is an implementation of the **Token Ring Algorithm** in Golang, to verify feasability for usage in multiple gateway support for CVMFS. The algorithm ensures mutual exclusion in a distributed system by passing a token among nodes in a logical ring structure. This test setup allows nodes to pass the token around, and communicate when a node fails.

## Features
- Implements a token-passing mechanism access to a common resource (`common.txt`).
- Uses HTTP for inter-node communication.
- Nodes can be appended to the ring while the other nodes are already active.
- Handles node failures by removing inactive nodes from the ring.
- Implements an acknowledgment system to confirm token receipt. If no acknowledgement is given within a timer, the token is instead passed to the next node. This prevents a deadlock when a failing node is given the token.

## Usage
### Running the nodes
1. Start docker
   ```sh
   docker-compose up --build
   ```