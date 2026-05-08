#!/bin/bash
# =============================================================================
# Sub2API Docker Deployment Preparation Script
# =============================================================================
# This script prepares deployment files for Sub2API:
#   - Downloads docker-compose.local.yml and .env.example
#   - Generates secure secrets (JWT_SECRET, TOTP_ENCRYPTION_KEY, POSTGRES_PASSWORD)
#   - Creates necessary data directories
#
# After running this script, you can start services with:
#   docker-compose up -d
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub repository info
GITHUB_REPO="qi758771475-gif/sub2api-main"
GITHUB_BRANCH="master"
GITHUB_DEPLOY_DIR="deploy"
GITHUB_RAW_URL="https://raw.githubusercontent.com/${GITHUB_REPO}/${GITHUB_BRANCH}/${GITHUB_DEPLOY_DIR}"
GITHUB_API_URL="https://api.github.com/repos/${GITHUB_REPO}/contents/${GITHUB_DEPLOY_DIR}"

# Download file from GitHub with fallback strategies
# Usage: github_download <filename> <output_path>
github_download() {
    local filename="$1"
    local output="$2"
    local raw_url="${GITHUB_RAW_URL}/${filename}"
    local api_url="${GITHUB_API_URL}/${filename}?ref=${GITHUB_BRANCH}"

    # Strategy 1: raw.githubusercontent.com (with proxy, i.e. default behavior)
    if command_exists curl; then
        curl -sSL --connect-timeout 10 --max-time 60 "${raw_url}" -o "${output}" 2>/dev/null
    elif command_exists wget; then
        wget -q --timeout=60 "${raw_url}" -O "${output}" 2>/dev/null
    fi

    # Validate: not empty and not an error page
    if [ -s "${output}" ] && ! grep -qi "404\|not found\|403\|forbidden" "${output}" 2>/dev/null; then
        return 0
    fi

    # Strategy 2: raw.githubusercontent.com without proxy
    if command_exists curl; then
        curl -sSL --noproxy '*' --connect-timeout 10 --max-time 60 "${raw_url}" -o "${output}" 2>/dev/null
    elif command_exists wget; then
        wget -q --no-proxy --timeout=60 "${raw_url}" -O "${output}" 2>/dev/null
    fi

    if [ -s "${output}" ] && ! grep -qi "404\|not found\|403\|forbidden" "${output}" 2>/dev/null; then
        return 0
    fi

    # Strategy 3: GitHub API (returns base64-encoded content in JSON)
    if command_exists curl; then
        local api_response
        api_response=$(curl -sSL --connect-timeout 10 --max-time 60 "${api_url}" 2>/dev/null)
        if [ -n "${api_response}" ] && echo "${api_response}" | grep -q '"content"'; then
            echo "${api_response}" | grep '"content"' | sed 's/.*"content"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' | tr -d '\n' | base64 -d > "${output}" 2>/dev/null
            if [ -s "${output}" ]; then
                return 0
            fi
        fi
    fi

    # All strategies failed
    return 1
}

# Print colored message
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Generate random secret
generate_secret() {
    openssl rand -hex 32
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main installation function
main() {
    echo ""
    echo "=========================================="
    echo "  Sub2API Deployment Preparation"
    echo "=========================================="
    echo ""

    # Check if openssl is available
    if ! command_exists openssl; then
        print_error "openssl is not installed. Please install openssl first."
        exit 1
    fi

    # Check if deployment already exists
    if [ -f "docker-compose.yml" ] && [ -f ".env" ]; then
        print_warning "Deployment files already exist in current directory."
        # Use /dev/tty for input when running via pipe (curl | bash)
        if [ -t 0 ]; then
            read -p "Overwrite existing files? (y/N): " -r
        elif [ -e /dev/tty ]; then
            read -p "Overwrite existing files? (y/N): " -r < /dev/tty
        else
            print_info "Non-interactive mode detected. Overwriting existing files."
            REPLY="y"
        fi
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Cancelled."
            exit 0
        fi
    fi

    # Download docker-compose.local.yml and save as docker-compose.yml
    print_info "Downloading docker-compose.yml..."
    if ! github_download "docker-compose.local.yml" "docker-compose.yml"; then
        print_error "Failed to download docker-compose.yml after multiple attempts."
        print_error "Please download manually from GitHub and place in current directory."
        exit 1
    fi
    print_success "Downloaded docker-compose.yml"

    # Download .env.example
    print_info "Downloading .env.example..."
    if ! github_download ".env.example" ".env.example"; then
        print_error "Failed to download .env.example after multiple attempts."
        print_error "Please download manually from GitHub and place in current directory."
        exit 1
    fi
    print_success "Downloaded .env.example"

    # Generate .env file with auto-generated secrets
    print_info "Generating secure secrets..."
    echo ""

    # Generate secrets
    JWT_SECRET=$(generate_secret)
    TOTP_ENCRYPTION_KEY=$(generate_secret)
    POSTGRES_PASSWORD=$(generate_secret)

    # Create .env from .env.example
    cp .env.example .env

    # Update .env with generated secrets (cross-platform compatible)
    if sed --version >/dev/null 2>&1; then
        # GNU sed (Linux)
        sed -i "s/^JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
        sed -i "s/^TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}/" .env
        sed -i "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
    else
        # BSD sed (macOS)
        sed -i '' "s/^JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
        sed -i '' "s/^TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}/" .env
        sed -i '' "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
    fi

    # Create data directories
    print_info "Creating data directories..."
    mkdir -p data postgres_data redis_data
    print_success "Created data directories"

    # Set secure permissions for .env file (readable/writable only by owner)
    chmod 600 .env
    echo ""

    # Display completion message
    echo "=========================================="
    echo "  Preparation Complete!"
    echo "=========================================="
    echo ""
    echo "Generated secure credentials:"
    echo "  POSTGRES_PASSWORD:     ${POSTGRES_PASSWORD}"
    echo "  JWT_SECRET:            ${JWT_SECRET}"
    echo "  TOTP_ENCRYPTION_KEY:   ${TOTP_ENCRYPTION_KEY}"
    echo ""
    print_warning "These credentials have been saved to .env file."
    print_warning "Please keep them secure and do not share publicly!"
    echo ""
    echo "Directory structure:"
    echo "  docker-compose.yml        - Docker Compose configuration"
    echo "  .env                      - Environment variables (generated secrets)"
    echo "  .env.example              - Example template (for reference)"
    echo "  data/                     - Application data (will be created on first run)"
    echo "  postgres_data/            - PostgreSQL data"
    echo "  redis_data/               - Redis data"
    echo ""
    echo "Next steps:"
    echo "  1. (Optional) Edit .env to customize configuration"
    echo "  2. Start services:"
    echo "     docker-compose up -d"
    echo ""
    echo "  3. View logs:"
    echo "     docker-compose logs -f sub2api"
    echo ""
    echo "  4. Access Web UI:"
    echo "     http://localhost:8080"
    echo ""
    print_info "If admin password is not set in .env, it will be auto-generated."
    print_info "Check logs for the generated admin password on first startup."
    echo ""
}

# Run main function
main "$@"
