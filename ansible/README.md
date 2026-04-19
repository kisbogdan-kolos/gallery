# Ansible deployment for Gallery app

## BME OKD

This playbook deploys the app to the BME OKD cluster. It needs the following environment variables set:

-    `openshift_namespace`
-    `app_postgres_pass`
-    `app_seaweed_access_key`
-    `app_seaweed_secret_key`
-    `app_jwt_secret`

It also needs a valid `.kubeconfig` file for the OpenShift cluster. This can be done using `oc login`, or using a service account inside GutHub actions.

For development, copy the login command from the web console, and log in using that. After this, the playbook can be run using Ansible.

You may need to install the `kubeconfig` python module, if not already present.

```bash
pip install kubeconfig
```

Then just run the playbook:

```bash
ansible-playbook bme-okd.yaml
```
