#!/bin/bash

# Aether Defense System Deployment Script
# Usage: ./scripts/deploy.sh [environment] [version]
# Example: ./scripts/deploy.sh dev latest

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check required tools
check_prerequisites() {
    print_info "Checking required tools..."

    if ! command -v helm &> /dev/null; then
        print_error "Helm is not installed, please install Helm first"
        exit 1
    fi

    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed, please install kubectl first"
        exit 1
    fi

    print_info "All required tools are installed"
}

# Show usage help
show_help() {
    echo "Aether Defense System Deployment Script"
    echo ""
    echo "Usage:"
    echo "  $0 [environment] [version] [options]"
    echo ""
    echo "Environment:"
    echo "  dev      - Development environment"
    echo "  staging  - Staging environment"
    echo "  prod     - Production environment"
    echo ""
    echo "Version:"
    echo "  latest   - Latest version"
    echo "  v1.0.0   - Specific version tag"
    echo "  abc1234  - Specific commit SHA"
    echo ""
    echo "Options:"
    echo "  --dry-run    - Only show commands to be executed, don't actually run"
    echo "  --help       - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 dev latest"
    echo "  $0 prod v1.0.0"
    echo "  $0 staging latest --dry-run"
}

# Parse arguments
ENVIRONMENT=""
VERSION=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_help
            exit 0
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            if [[ -z "$ENVIRONMENT" ]]; then
                ENVIRONMENT=$1
            elif [[ -z "$VERSION" ]]; then
                VERSION=$1
            else
                print_error "Unknown argument: $1"
                show_help
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate arguments
if [[ -z "$ENVIRONMENT" ]]; then
    print_error "Please specify environment (dev/staging/prod)"
    show_help
    exit 1
fi

if [[ -z "$VERSION" ]]; then
    print_error "Please specify version"
    show_help
    exit 1
fi

# Validate environment
case $ENVIRONMENT in
    dev|staging|prod)
        ;;
    *)
        print_error "Invalid environment: $ENVIRONMENT (supported: dev/staging/prod)"
        exit 1
        ;;
esac

# Set environment variables
NAMESPACE="aether-defense-${ENVIRONMENT}"
RELEASE_NAME="aether-defense-${ENVIRONMENT}"
VALUES_FILE="deploy/helm/aether-defense/values-${ENVIRONMENT}.yaml"
CHART_PATH="deploy/helm/aether-defense"

# Check if values file exists
if [[ ! -f "$VALUES_FILE" ]]; then
    print_warn "Environment config file not found: $VALUES_FILE, using default config"
    VALUES_FILE="deploy/helm/aether-defense/values.yaml"
fi

# Build Helm command
HELM_CMD="helm upgrade --install $RELEASE_NAME $CHART_PATH"
HELM_CMD="$HELM_CMD --namespace $NAMESPACE"
HELM_CMD="$HELM_CMD --create-namespace"
HELM_CMD="$HELM_CMD --values $VALUES_FILE"

# Set image versions
REGISTRY="ghcr.io/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\)\.git/\1/' | tr '[:upper:]' '[:lower:]')"
HELM_CMD="$HELM_CMD --set global.imageRegistry=$REGISTRY"
HELM_CMD="$HELM_CMD --set userApi.image.tag=$VERSION"
HELM_CMD="$HELM_CMD --set userRpc.image.tag=$VERSION"
HELM_CMD="$HELM_CMD --set tradeRpc.image.tag=$VERSION"
HELM_CMD="$HELM_CMD --set promotionRpc.image.tag=$VERSION"

# Add timeout and wait
HELM_CMD="$HELM_CMD --wait --timeout=10m"

# Show deployment information
print_info "Preparing to deploy Aether Defense System"
print_info "Environment: $ENVIRONMENT"
print_info "Version: $VERSION"
print_info "Namespace: $NAMESPACE"
print_info "Image registry: $REGISTRY"
print_info "Config file: $VALUES_FILE"

if [[ "$DRY_RUN" == "true" ]]; then
    print_warn "DRY RUN mode - showing commands only, not executing"
    echo ""
    echo "Commands to be executed:"
    echo "$HELM_CMD"
    exit 0
fi

# Confirm deployment
if [[ "$ENVIRONMENT" == "prod" ]]; then
    print_warn "You are about to deploy to production!"
    read -p "Confirm to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Deployment cancelled"
        exit 0
    fi
fi

# Check prerequisites
check_prerequisites

# Check kubectl connection
print_info "Checking Kubernetes connection..."
if ! kubectl cluster-info &> /dev/null; then
    print_error "Cannot connect to Kubernetes cluster"
    exit 1
fi

# Execute deployment
print_info "Starting deployment..."
echo "Executing command: $HELM_CMD"
echo ""

if eval "$HELM_CMD"; then
    print_info "Deployment completed successfully!"

    # Show deployment status
    print_info "Checking deployment status..."
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME"

    # Show service information
    print_info "Service information:"
    kubectl get svc -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME"

    print_info "Deployment completed!"
    print_info "You can monitor the application with these commands:"
    echo "  kubectl get pods -n $NAMESPACE"
    echo "  kubectl logs -f deployment/$RELEASE_NAME-user-api -n $NAMESPACE"

else
    print_error "Deployment failed!"
    print_info "Checking failed pods:"
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$RELEASE_NAME"

    print_info "View detailed error information:"
    echo "  kubectl describe pods -n $NAMESPACE"
    echo "  kubectl logs -n $NAMESPACE -l app.kubernetes.io/instance=$RELEASE_NAME"

    exit 1
fi
