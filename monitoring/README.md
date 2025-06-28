# Monitoring Setup

This directory contains the monitoring configuration for the image processing application.

## Components

### Prometheus
- **Configuration**: `prometheus.yml`
- **Port**: 9090
- **Purpose**: Collects and stores metrics from all services

### Grafana
- **Port**: 3000
- **Default Credentials**: admin/admin
- **Purpose**: Visualizes metrics and provides dashboards

### PostgreSQL Monitoring

#### PostgreSQL Exporter
- **Image**: `prometheuscommunity/postgres-exporter:latest`
- **Port**: 9187
- **Purpose**: Exports PostgreSQL metrics to Prometheus
- **Configuration**: Uses custom queries from `postgres-queries.yaml`

#### Custom Queries
The `postgres-queries.yaml` file defines custom queries to collect detailed PostgreSQL metrics:

- **Database Statistics**: Connections, transactions, row operations
- **Buffer Cache**: Hit ratios and performance metrics
- **Connection States**: Active, idle, and other connection states
- **Locks**: Database lock monitoring
- **Background Writer**: Checkpoint and buffer statistics
- **Database Size**: Size monitoring for all databases

#### Grafana Dashboard
The `postgresql-dashboard.json` provides a comprehensive dashboard with:

1. **Active Connections**: Real-time connection count
2. **Transaction Rate**: Commits and rollbacks per second
3. **Row Operations Rate**: Insert, update, delete operations
4. **Buffer Cache Hit Ratio**: Performance indicator
5. **Connection States**: Breakdown of connection states
6. **Database Locks**: Lock monitoring
7. **Database Size**: Size tracking over time
8. **Deadlocks Rate**: Deadlock monitoring

## Access URLs

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **PostgreSQL Exporter**: http://localhost:9187/metrics

## Metrics Available

### Key PostgreSQL Metrics
- `pg_stat_database_numbackends`: Active connections
- `pg_stat_database_xact_commit`: Committed transactions
- `pg_stat_database_xact_rollback`: Rolled back transactions
- `pg_stat_database_tup_*`: Row operations (inserted, updated, deleted)
- `pg_stat_database_blks_*`: Buffer cache statistics
- `pg_stat_database_deadlocks`: Deadlock count
- `pg_stat_activity_count`: Connection states
- `pg_locks_count`: Database locks
- `pg_database_size_bytes`: Database sizes

## Troubleshooting

1. **PostgreSQL Exporter not connecting**: Check if PostgreSQL is healthy and accessible
2. **No metrics in Prometheus**: Verify the postgres-exporter target is up in Prometheus targets
3. **Dashboard not loading**: Ensure Prometheus datasource is configured in Grafana

## Customization

To add more metrics:
1. Add queries to `postgres-queries.yaml`
2. Update the Prometheus configuration if needed
3. Add panels to the Grafana dashboard 