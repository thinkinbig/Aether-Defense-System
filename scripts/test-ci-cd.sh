#!/bin/bash

# Test CI/CD Pipeline Script
# This script helps you test the CI/CD pipeline step by step

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
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

# Check if we're in a git repository
check_git_repo() {
    print_step "Checking if we're in a git repository..."
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository. Please run this from your project root."
        exit 1
    fi
    print_success "Git repository detected"
}

# Check if GitHub remote exists
check_github_remote() {
    print_step "Checking GitHub remote..."
    if ! git remote get-url origin | grep -q "github.com"; then
        print_error "No GitHub remote found. Please add GitHub as origin remote."
        exit 1
    fi

    REPO_URL=$(git remote get-url origin)
    print_success "GitHub remote found: $REPO_URL"
}

# Check if workflows exist
check_workflows() {
    print_step "Checking workflow files..."

    if [ ! -f ".github/workflows/ci.yml" ]; then
        print_error "CI workflow not found at .github/workflows/ci.yml"
        exit 1
    fi
    print_success "CI workflow found"

    if [ ! -f ".github/workflows/cd.yml" ]; then
        print_error "CD workflow not found at .github/workflows/cd.yml"
        exit 1
    fi
    print_success "CD workflow found"
}

# Show next steps
show_next_steps() {
    echo ""
    echo -e "${BLUE}=== NEXT STEPS ===${NC}"
    echo ""
    echo "1. Push your code to GitHub:"
    echo "   git add ."
    echo "   git commit -m 'Add CI/CD pipeline'"
    echo "   git push origin main"
    echo ""
    echo "2. Check GitHub Actions:"
    echo "   - Go to your repository on GitHub"
    echo "   - Click 'Actions' tab"
    echo "   - You should see CI workflow running"
    echo ""
    echo "3. Test CD pipeline:"
    echo "   - Push to 'develop' branch to test dev deployment"
    echo "   - Create a tag 'v1.0.0' to test staging/prod deployment"
    echo ""
    echo "4. Monitor container registry:"
    echo "   - Go to your GitHub profile → Packages"
    echo "   - You should see Docker images after successful builds"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}=== CI/CD Pipeline Test ===${NC}"
    echo ""

    check_git_repo
    check_github_remote
    check_workflows

    print_success "All checks passed! 🎉"
    show_next_steps
}

# Run main function
main "$@"
