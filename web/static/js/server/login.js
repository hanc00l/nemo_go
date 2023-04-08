const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5KPrGCKxttISC5hFPy44tt57zzYhIrLcUet354BLWImPZXY2TaqWoirLsWhbkFaYNu6U0du0Rcgj5RBjFF8GvhxDvQE1vG8JP/g8uDkhZSWoYhHyZrvvcTTKbmZESc6xFnRY5i3cWFDMWxAt+sRaZd/Yr3vw2VPZYZpcuL+3fTQEEAdC2roFephjN3rXDuj/Ern7hbD/+q2QcOUqJ6XwnOO+9lBX9ID1LK0JINVxh4/YaXe5RHBOR0WkWgPDz6IOFGtVkA1ACTiHuU/bMKMbhXXN1Nh93JgdGodhP7ac/tzIH9574ardrlg+cIA6zWNEEm3K+CFDaaCLWm5JKClHbwIDAQAB-----END RSA Public Key-----';
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