package mysql

// Preserves non-parseable lines (assuming they are comments).
const doComments = true

// Preserves INSERT statements (disable to make debug printfs readable).
const doInserts = true

// Preserves non-insert lines (LOCK/UNLOCK/SET/...).
const doMisc = true
