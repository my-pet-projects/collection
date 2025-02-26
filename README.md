# collection

## Useful database scripts

### Create dump

```bash
turso db shell beer-collection-fresh-test .dump > beer-collection-dump.sql
```

### Import from dump

```bash
turso db shell geography < ./dumps/geography-dump.sql
turso db shell beer-collection < ./dumps/beer-collection-dump.sql
```
