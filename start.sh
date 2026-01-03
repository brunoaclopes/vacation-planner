#!/bin/bash

# Start Backend
echo "Starting Go backend..."
cd backend
go mod tidy
go run cmd/server/main.go &
BACKEND_PID=$!

# Wait for backend to start
sleep 3

# Start Frontend
echo "Starting React frontend..."
cd ../frontend
npm install
npm run dev &
FRONTEND_PID=$!

echo ""
echo "============================================"
echo "Portuguese Vacation Planner is running!"
echo "============================================"
echo "Frontend: http://localhost:5173"
echo "Backend:  http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop both servers"
echo "============================================"

# Wait for any key to stop
wait $BACKEND_PID $FRONTEND_PID
