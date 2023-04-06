const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5thHdelU5JOAmb+1BW05oCUFkpZBUWtl4SXfbTE9D7Lo6oRGGTdHQrZKNbtGc5wAkrt+6qHT4xs1CjVn0V4n9zY1WaIMkQiyF0sMKy1TJz7A9dJ7W9bTiBr9wVCCB+63wAc8pxFcWCKD1YWXJr2x/1sJu7eRwFc1hLPnOdUiDb1I592CIdh+0YXsyEO5KDD59R8/u2DvzFF9pAlipAgm4BCZYvmVz021kAX7NX0HbD/qJOGwzqiTsUdVdsUh2jqpY2pA8T3z4pIbXhw0eufpOtu55qFpW1xzrij2UnXoj3zpXRrtaP34Y6XkcD6UP4x8neRdUhmrD1W6g/CQNEppAQIDAQAB-----END RSA Public Key-----';
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