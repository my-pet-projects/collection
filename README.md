# collection

## Useful database scripts

### Create dump

```bash
turso auth login
turso db shell beer-collection .dump > ./dumps/beer-collection-dump.sql
```

### Import from dump

```bash
turso db create beer-collection
turso db shell geography < ./dumps/geography-dump.sql
turso db shell beer-collection < ./dumps/beer-collection-dump.sql
turso db tokens create beer-collection
```
