#!/bin/sh

/Users/scott/Dev/abc/abc -c 'pdr -T 300 -v' $1 > bench/abc/`basename $1`.out

