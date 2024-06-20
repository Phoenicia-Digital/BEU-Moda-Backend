# Dockerfile
FROM postgres:16

# Set environment variables
ENV POSTGRES_USER=phoeniciadigital
ENV POSTGRES_PASSWORD=pdsoftware
ENV POSTGRES_DB=pd_database

# Copy initialization script
COPY sql/init.sql /docker-entrypoint-initdb.d/
