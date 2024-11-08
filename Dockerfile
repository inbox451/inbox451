FROM chainguard/wolfi-base:latest

# Set the working directory
WORKDIR /inbox451

# Copy only the necessary files
COPY inbox451 .
COPY config.yml.sample config.yml

# Expose the application ports
# 8080: HTTP
# 1025: SMTP
# 1143: IMAP
EXPOSE 8080 1025 1143

# Define the command to run the application
CMD ["./inbox451"]
