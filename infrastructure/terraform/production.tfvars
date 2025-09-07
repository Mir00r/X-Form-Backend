# Production Environment Configuration

# Basic settings
environment = "production"
aws_region  = "us-west-2"

# Database settings (production-ready)
db_instance_class        = "db.r6g.large"
db_allocated_storage     = 100
db_max_allocated_storage = 1000

# Redis settings (production cluster)
redis_node_type        = "cache.r6g.large"
redis_num_cache_nodes  = 3

# Cost optimization (no spot instances for production)
enable_spot_instances = false

# Monitoring (comprehensive for production)
log_retention_days = 30
backup_retention_period = 7
enable_point_in_time_recovery = true

# Auto scaling
auto_scaling_enabled = true

# Observability stack
enable_observability_stack = true

# Domain configuration (update with your domain)
domain_name = "api.xform.example.com"
certificate_arn = "" # Add your SSL certificate ARN

# Security (restrictive for production)
# allowed_cidr_blocks = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"] # Private networks only
allowed_cidr_blocks = ["0.0.0.0/0"] # Update this with your specific IP ranges

# VPC Flow Logs
enable_vpc_flow_logs = true
