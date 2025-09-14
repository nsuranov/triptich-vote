#!/bin/bash

for ((i=1; i<=150000000; i++)); do
  rnd=$(openssl rand -hex 8)
  go run ../cmd/keygen "$rnd" http://localhost:8086
  rm "$rnd"-key.json
done
