#!/bin/bash

# QuotaSense CLI Release Script
# Usage: ./release.sh <version> [notes]

VERSION=$1
NOTES=$2

# Auto-detect version if not provided
if [ -z "$VERSION" ]; then
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
    if [ -z "$LATEST_TAG" ]; then
        VERSION="v0.1.0"
        echo "No tags found. Starting with $VERSION"
    else
        echo "Latest tag found: $LATEST_TAG"
        # Increment patch version (assuming semver vX.Y.Z)
        if [[ $LATEST_TAG =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
            major="${BASH_REMATCH[1]}"
            minor="${BASH_REMATCH[2]}"
            patch="${BASH_REMATCH[3]}"
            new_patch=$((patch + 1))
            VERSION="v${major}.${minor}.${new_patch}"
            echo "Auto-incremented version to $VERSION"
        else
            echo "Warning: Latest tag $LATEST_TAG does not match vX.Y.Z format."
            echo "Usage: ./release.sh <version> [notes]"
            exit 1
        fi
    fi
fi

# Ensure version starts with 'v'
if [[ ! $VERSION == v* ]]; then
    VERSION="v$VERSION"
fi

BINARY_NAME="qs"
DIST_DIR="dist"

echo "🚀 Preparing release $VERSION..."

# Update version in cmd/version.go
sed -i '' "s/var Version = \".*\"/var Version = \"$VERSION\"/" cmd/version.go 2>/dev/null || \
sed -i "s/var Version = \".*\"/var Version = \"$VERSION\"/" cmd/version.go

# Clean dist directory
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Build for different platforms
platforms=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    PLATFORM_SPLIT=(${platform//\// })
    GOOS=${PLATFORM_SPLIT[0]}
    GOARCH=${PLATFORM_SPLIT[1]}

    OUTPUT_NAME=$BINARY_NAME
    if [ $GOOS = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo "📦 Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -o "$DIST_DIR/$OUTPUT_NAME" main.go

    # Package
    PACKAGE_NAME="${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}"
    if [ $GOOS = "windows" ]; then
        zip -j "$DIST_DIR/${PACKAGE_NAME}.zip" "$DIST_DIR/$OUTPUT_NAME" > /dev/null
    else
        tar -czf "$DIST_DIR/${PACKAGE_NAME}.tar.gz" -C $DIST_DIR $OUTPUT_NAME
    fi

    rm "$DIST_DIR/$OUTPUT_NAME"
done

echo "📝 Generating changelog..."
if [ -z "$NOTES" ]; then
    # Get changes since last tag
    PREVIOUS_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
    if [ -z "$PREVIOUS_TAG" ]; then
        NOTES=$(git log --oneline | head -n 10)
    else
        NOTES=$(git log $PREVIOUS_TAG..HEAD --oneline)
    fi
fi

echo "📤 Uploading to GitHub..."
gh release create "$VERSION" $DIST_DIR/*.tar.gz $DIST_DIR/*.zip --title "Release $VERSION" --notes "$NOTES"

echo "✅ Release $VERSION completed successfully!"
