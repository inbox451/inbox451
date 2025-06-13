FROM chainguard/wolfi-base:latest

# Add non-root user
USER nonroot

# Create necessary directories with correct permissions
WORKDIR /inbox451

# Default environment variables
ENV INBOX451_SERVER_HTTP_PORT=":8080" \
    INBOX451_SERVER_SMTP_PORT=":1025" \
    INBOX451_SERVER_IMAP_PORT=":1143" \
    INBOX451_LOGGING_LEVEL="info" \
    INBOX451_LOGGING_FORMAT="json"

# Copy binary
COPY --chown=nonroot:nonroot inbox451 .

# Expose ports
EXPOSE 8080 1025 1143

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost${INBOX451_SERVER_HTTP_PORT}/api/health || exit 1

# Run the application
ENTRYPOINT ["./inbox451"]
