# Local Database

MSA uses SQLite (via [GORM](https://github.com/go-gorm/gorm)) for local data storage.

## What's Stored

- **Accounts**: User accounts with balance tracking
- **Positions**: Stock holdings with cost basis and P&L
- **Transactions**: Buy/sell orders with status tracking

## Database Location

```
~/.msa/msa.sqlite
```

## Backup & Restore

```bash
# Backup
cp ~/.msa/msa.sqlite ~/.msa/msa.sqlite.backup.$(date +%Y%m%d)

# Restore
cp ~/.msa/msa.sqlite.backup.YYYYMMDD ~/.msa/msa.sqlite
```

## Amount Units

All monetary amounts are stored as integers in "毫" (1 Yuan = 10000 毫) to avoid floating-point precision issues.

- `10000` = 1.00 元
- Display: `amount / 10000 = displayed value`
