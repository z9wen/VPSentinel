CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    agent_version TEXT NOT NULL,
    os_name TEXT,
    kernel_version TEXT,
    last_seen_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE metric_batches (
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    batch_id TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    accepted_samples INTEGER NOT NULL CHECK (accepted_samples >= 0),
    PRIMARY KEY (node_id, batch_id)
);

CREATE TABLE metric_samples (
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    sample_id TEXT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    cpu_percent DOUBLE PRECISION CHECK (cpu_percent BETWEEN 0 AND 100),
    memory_used_bytes BIGINT CHECK (memory_used_bytes >= 0),
    memory_total_bytes BIGINT CHECK (memory_total_bytes >= 0),
    disk_used_bytes BIGINT CHECK (disk_used_bytes >= 0),
    disk_total_bytes BIGINT CHECK (disk_total_bytes >= 0),
    load_1 DOUBLE PRECISION CHECK (load_1 >= 0),
    load_5 DOUBLE PRECISION CHECK (load_5 >= 0),
    load_15 DOUBLE PRECISION CHECK (load_15 >= 0),
    uptime_seconds BIGINT CHECK (uptime_seconds >= 0),
    network_rx_bytes BIGINT CHECK (network_rx_bytes >= 0),
    network_tx_bytes BIGINT CHECK (network_tx_bytes >= 0),
    failed_collectors TEXT[] NOT NULL DEFAULT '{}',
    PRIMARY KEY (node_id, sample_id)
);

CREATE INDEX metric_samples_node_recorded_at_idx
    ON metric_samples (node_id, recorded_at DESC);

CREATE INDEX metric_samples_recorded_at_idx
    ON metric_samples (recorded_at);

CREATE INDEX nodes_last_seen_at_idx
    ON nodes (last_seen_at DESC);
