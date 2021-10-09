base release, you can use this version in your test env。

In this version，deployment stroage is unsafe and trino metrics Unrealized .

### version 0.1
- [X] Basis function  
- [X] add a trino cluster in kubernetes 
- [X] edit crds to if you change config, restart the necessary parts
- [X] add or delete worker when you just change the num 
- [X] watch deployment and pod, logs trino coordinator and worker stats in **trinos** status items



In next version deployment to statefulset and add more settings,
and the trino metrics with prometheus will be added.

### In next  0.2 version
- [ ] change deploy to stateful and add more settings 
- [ ] add metrics, use prometheus or others

### Not planning
- [ ] worker auto discovery coordinator without config