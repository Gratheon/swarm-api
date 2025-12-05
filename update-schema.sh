#!/bin/bash

# This script updates the GraphQL schema in the schema registry after schema changes

echo "🔄 Updating GraphQL Schema in Registry..."
echo ""

# Step 1: Generate version file
echo "1️⃣  Generating version from git commit..."
cd /Users/artjom/git/swarm-api
git rev-parse --short HEAD > .version
VERSION=$(cat .version)
echo "   Version: $VERSION"
echo ""

# Step 2: Regenerate GraphQL code (if needed)
echo "2️⃣  Regenerating GraphQL code..."
just gen
echo ""

# Step 3: Restart swarm-api to push new schema
echo "3️⃣  To update the schema registry, you need to restart swarm-api:"
echo ""
echo "   Option A - If running with 'just develop':"
echo "   • Stop the current process (Ctrl+C)"
echo "   • Run: cd /Users/artjom/git/swarm-api && just develop"
echo ""
echo "   Option B - If running with docker-compose:"
echo "   • Run: cd /Users/artjom/git/swarm-api && just stop && just start"
echo ""
echo "   Option C - Quick restart (if already running):"
echo "   • The schema will auto-register on service startup"
echo ""

echo "✅ Version file updated. Restart swarm-api to push new schema!"
echo ""
echo "📝 New fields added to Hive type:"
echo "   - parentHive: Hive"
echo "   - childHives: [Hive]"
echo "   - splitDate: DateTime"
echo ""
echo "📝 New mutation added:"
echo "   - splitHive(sourceHiveId: ID!, name: String!, frameIds: [ID!]!): Hive"

