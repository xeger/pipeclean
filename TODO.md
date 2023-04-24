1. supress mysql unique string indexes in mark mode (or always?)
    - e.g. UNIQUE KEY `index_users_on_email` (`email`),
1. add nickname/friendly-name to default policy
1. Heuristic policy (e.g. base64):
    - allow tunable confidence
1. parse CSV string columns & sanitize the bits
1. Functional tests
    - need dummy schema + data set
    - need lots of data variability
