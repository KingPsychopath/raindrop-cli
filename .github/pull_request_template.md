## Summary

- 

## Validation

- [ ] `make check`
- [ ] `make check-integrations` if adapters or packaging changed
- [ ] README/docs updated when command behavior changed
- [ ] No tokens, exports, local `.env` files, or binaries committed

- [ ] CLI help and `docs/COMMANDS.md` agree

## Safety

- [ ] Destructive behavior is explicit
- [ ] Bulk mutation has a dry-run or review path when appropriate
- [ ] Model inputs preserve library vocabulary; uncertainty is withheld from apply
- [ ] Tool intent and host authorization are documented separately
