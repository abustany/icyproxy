#!/bin/bash

UNFORMATTED_FILES=$(gofmt -l .)

if [ -n "$UNFORMATTED_FILES" ]; then
  echo "The following files are not formatted:"
  echo "$UNFORMATTED_FILES"
  exit 1
fi
