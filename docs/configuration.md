# Configuration

MSA supports multiple configuration methods with the following priority (high to low): **CLI parameters > Environment variables > Config file > Default values**

## Interactive Configuration

```bash
msa config
```

In the TUI configuration interface, you can:
- Use arrow keys to select configuration items
- Press `Enter` to edit
- Auto-fill Base URL when selecting Provider
- Press `S` to save
- Press `R` to reset to defaults
- Press `Q` to quit

## Environment Variables

```bash
export MSA_PROVIDER=siliconflow
export MSA_API_KEY=sk-xxxxxxxxxxxx
export MSA_BASE_URL=https://api.example.com/v1       # optional
export MSA_LOG_LEVEL=debug                            # optional
export MSA_LOG_FILE=/path/to/msa.log                  # optional
```

## CLI Parameters

| Parameter | Short | Description | Example |
|-----------|-------|-------------|---------|
| `--question` | `-q` | Single-round question (no TUI) | `msa -q "Tencent stock code?"` |
| `--model` | `-m` | Override model for this run | `-m "deepseek-r1"` |
| `--resume` | - | Resume a previous session by ID | `--resume 2026-01-01_uuid` |
| `--config` | - | Set config (file or key=value) | `--config apikey=sk-xxx` |

```bash
# Resume a previous session
msa --resume <session-id>

# Use config file
msa --config /path/to/config.json chat

# Use key=value format
msa --config apikey=sk-xxx --config loglevel=debug chat

# Mixed usage
msa --config /path/to/config.json --config apikey=sk-xxx chat
```

## View Current Configuration

In the chat interface:

```
/config
```

This displays: Provider, Model, Base URL, API Key (partially hidden), log level, and log file path.

## Configuration File

Configuration is saved at `~/.msa/msa_config.json`:

```json
{
  "provider": "siliconflow",
  "model": "deepseek-ai/DeepSeek-R1",
  "base_url": "https://api.siliconflow.cn/v1",
  "api_key": "sk-xxxxxxxx",
  "log_config": {
    "level": "debug",
    "format": "json",
    "output": "stderr",
    "file_path": "/tmp/msa.log"
  }
}
```

## Security

API Keys are stored in plaintext in the configuration file. It is recommended to:

```bash
chmod 600 ~/.msa/msa_config.json
chmod 700 ~/.msa/
```

- Do not commit configuration files to version control
- Do not share configurations containing API Keys
- Rotate API Keys regularly
