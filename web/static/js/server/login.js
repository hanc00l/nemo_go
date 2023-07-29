const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0yIxvD9yntuKAuwqagWMDQX2flAV14WltxllK30EHPJNpyNXqYzWarqYg3nFh2nII5ItSxUkBCNhXE7qFUl6TfLwljqhChOvBQE7huwFFBvAOk+9Vr7OrJ5ht2QA4yrBogBRZDlmLV1tV5KUoEgBXFe8kSeRffZGjnQv6REZzOCXiXswuSdKAf548AvbdyaD8xZcbQ5wfH+vYQlKJV4Tt7Qd8dqNt/u+OBoKXEP4hBzZ/zHmgIyLB6ouoNh9zYAih9Z+X1HoTPcnVUQBVob3Ijq+7RiEzkGYZ/ynxvGc5e1yOyGcbyluNE+mV7ScZkTiMTmiJfZShzvcQDX6OC4roQIDAQAB-----END RSA Public Key-----';
$(function () {
    $("#button_login").click(function () {
        let username = $('#username').val();
        let password = $('#password').val();
        if (username === "" || password === "") {
            return false;
        }
        let encryptor = new JSEncrypt();
        encryptor.setPublicKey(pubKey);
        let rsaUsername = encryptor.encrypt(username);
        let rsaPassword = encryptor.encrypt(password);
        $('#username').val(rsaUsername);
        $('#password').val(rsaPassword);

        return true;
    });
});