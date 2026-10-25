#!/bin/bash

ps aux | grep "trpc_main" | grep -v grep | awk '{print $2}' | xargs kill -9
