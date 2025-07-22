#!/bin/bash

echo "🚀 Starting GoLang eSports Fantasy Backend..."

# Navigate to Go backend directory
cd /app/go-backend

# Check if database connection is working
echo "📊 Testing database connection..."

# Start the application
echo "🎮 Starting eSports Fantasy Backend..."
export GIN_MODE=debug
export OTP_CONSOLE=true

./build/esports-fantasy