const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvo263Ie8d/luLRz2a4p1KxuLVQZSSK8xH2C+VlVdq+S8bsoC1+JlunFlFPO+w+IdIYBKhBR27/O+2vCm1MpOPiGGXCeFBNBgA8gKGyH06jhJ9OAXnDSjs4h0GcHAMETUQvAL52NVHcUuPyror3DQvv6TiiR5xwIBEFGrfvyOxBpaFUz+fbLuaEsrHETlO+tYD6O64DAf4pghJDEOBp4yV4zhY1M561Qty023Q741vD8z1gQ3DdJnHpSw0I+L30TpwqbPKb9g/62EoOyw0QywhCZyXla9bkeEpJ8OgX2sClKU1FT1a+pUgrSq7xRyuLTJNilzV0z6BGJx8XZK9h2QYQIDAQAB-----END RSA Public Key-----';
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