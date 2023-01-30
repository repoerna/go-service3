# DEPLOYMENT
0. make kind-down # delete cluster
1. make all # build services images
2. make kind-up  # run kind cluster
3. make king-load # run container inside cluster