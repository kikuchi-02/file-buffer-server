#!/bin/bash
for i in {0..100}
do
    curl -d '{"price": 1, "origin": "tokyo", "kind": "シナノゴールド"}' -X POST localhost:8000;
done