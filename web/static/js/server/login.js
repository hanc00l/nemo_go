const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt/LHWuvcmTVjTLzWY2kirzs+XRMjWKntvBev85bDAAVmd+6O3D1cBKd5bqW0T8mKmX7c1sxeFdK1VLntRCoZo4XttfRgoSl1XIWRwCxht83IMPC2Au0Nfc9lkC4xI1sKA2deVm7a8T0MX17VnMomtDPcIX9yBLG1BVJmL2+g8IYV+Vs2ND/fpBkxVGEWQ1kqUJelnVJ74dkkxUd0rhwcQGizuFbJLjGBeysyg+jamwiybG/rTkkIfuEmFGN/Z4tMqfe/8qlHIwGbfDgFx/94dPHDGUvVA3cBGZXTxDh80U8pWD6po1bdisNJZx9xcWTjokIJTJxBdx0QCoSBK4ZQoQIDAQAB-----END RSA Public Key-----';
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