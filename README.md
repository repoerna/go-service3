# Makefile
- make kind-down # delete cluster
- make all # build services images
- make kind-up  # run kind cluster
- make kind-load # run container inside cluster
- make kind-apply # re-run pod if any change in deployment file
- make kind-restart # restart pod
- make kind-update # rebuild app, load updated image and restart pod
- make kind-update-apply # rebuild app, load and apply deployment configuration
