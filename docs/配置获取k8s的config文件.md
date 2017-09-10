//设置k8s的集群
kubectl config set-cluster k8s-cluster --server=http://apiserverIp:8080 --api-version=v1

//设置k8s的链接的用户信息
kubectl config set-credentials myself --username=admin --password=secret

//设置集群的默认context
kubectl config set-context default-context --cluster=k8s-cluster --user=myself

//设置默认context
kubectl config use-context default-context

//设置context下的默认namespace
kubectl config set contexts.default-context.namespace default

然后执行命令：kubectl config view