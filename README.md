# calibre-browser

Simple CLI and web-based calibre collection browser

## Project Structure

- `main.go` - Main application entry point with unified CLI and web server
- `internal/` - Internal packages for database, handlers, and utilities
- `templates/` - HTML templates for web interface
- `static/` - Static assets
- `queries/` - SQL queries for database operations
- `schemas/` - Database schema definitions
- `deprecated/` - **Legacy code no longer in active use**
  - Contains previous client/server implementations that have been superseded by the current unified architecture
  - Files are preserved for reference and potential future use
  - Not part of the active build process
