package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// ANSI escape codes for background colors
const (
	ColorReset      = "\033[0m"
	BackgroundGreen = "\033[42m" // Success
	BackgroundRed   = "\033[41m" // Error
	BackgroundBlue  = "\033[44m" // Method
)

func StructuredLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Log start time and method
		start := time.Now()
		method := info.FullMethod

		// Get client IP address
		p, ok := peer.FromContext(ctx)
		var clientIP string
		if ok {
			clientIP, _, _ = net.SplitHostPort(p.Addr.String())
		} else {
			clientIP = "unknown"
		}

		// Handle the request
		resp, err := handler(ctx, req)

		// Log response status and duration with background color coding
		duration := time.Since(start).Milliseconds()
		status := "Success"            // Default status
		statusColor := BackgroundGreen // Default color
		if err != nil {
			status = "Failed "
			statusColor = BackgroundRed
		}

		// Format log output
		methodField := fmt.Sprintf("%-40s", method)     // Align method column
		statusField := fmt.Sprintf("%-7s", status)      // Align status column
		durationField := fmt.Sprintf("%4dms", duration) // Align duration column

		log.Printf("%s %s %s | %s | %s |%s %s %s",
			statusColor,
			statusField,
			ColorReset,
			durationField,
			clientIP,
			BackgroundBlue,
			methodField,
			ColorReset,
		)

		return resp, err
	}
}
