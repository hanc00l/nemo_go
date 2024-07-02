const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA433Xb1ineqNT/id4EPq3t7w+oH8imDPjtFl69guwuYU0aQaKqwdYsB3i4vRXUv3wlum88DZBV3cnbpCelU3e0iSyququ+oaxlDBjcbC3m288veZVeooJmzBCLsKimVDmCLRud8LUR5VxGaBU7EgpnVRYLkd65um8FH63F7yBdtdmuFkkOXvx8lcxIrV7Rn8ONpXdaXZs6JzfF/Zq1DPJpQqVfzmwLhFj9JVf/EZM79gbYYw96WDr5Ch06LKD8sqArEHOu/FbQlEpH1E6LF8dJGDepxD/NWvZhPwI7O9VCwosCRuEoXXp3mLT2JY9AEFSAWucERS8ybEQansQvxs5HQIDAQAB-----END RSA Public Key-----';
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