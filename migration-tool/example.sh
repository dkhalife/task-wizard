#!/bin/bash
# Simple example script demonstrating the migration tool usage
# This is for demonstration purposes only

set -e

echo "======================================"
echo "Task Wizard Migration Tool - Example"
echo "======================================"
echo ""
echo "This script demonstrates how to use the migration tool."
echo ""
echo "Prerequisites:"
echo "  1. SQLite database file with task wizard data"
echo "  2. MariaDB server running and accessible"
echo "  3. MariaDB database created with proper schema"
echo ""
echo "Example command:"
echo ""
echo "./migrate \\"
echo "  --sqlite /path/to/task-wizard.db \\"
echo "  --maria-host localhost \\"
echo "  --maria-port 3306 \\"
echo "  --maria-db taskwizard \\"
echo "  --maria-user taskuser \\"
echo "  --maria-pass taskpass"
echo ""
echo "For more information, see README.md"
