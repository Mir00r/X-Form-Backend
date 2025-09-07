# Development Environment Configuration

# Basic settings
environment = "dev"
aws_region  = "us-west-2"

# Database settings (smaller for dev)
db_instance_class        = "db.t3.micro"
db_allocated_storage     = 20
db_max_allocated_storage = 50

# Redis settings (smaller for dev)
redis_node_type        = "cache.t3.micro"
redis_num_cache_nodes  = 1

# Cost optimization for dev
enable_spot_instances = true

# Monitoring (basic for dev)
log_retention_days = 7
backup_retention_period = 3

# Auto scaling
auto_scaling_enabled = false

# Observability stack
enable_observability_stack = true

# Domain (optional for dev)
domain_name = ""
certificate_arn = ""

# Security (more permissive for dev)
allowed_cidr_blocks = ["0.0.0.0/0"]
