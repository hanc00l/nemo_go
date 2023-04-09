const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArmv0STPCPpw9Z99d2IPAUTnag20sTHkurgv3fWa6JBXIP3kjBxg7E75nLQLfuXff9VqdmuIUC3qwy5Z69uVhTyb5JdPC2uGd0oZ/wDniCXy8C4JQwSE6PmgsjXvgE3aAVhHJTUgBzElMn0FvEl95F6owJJBA6f77FCtTQvDDBWxsRXBLE+zkP32ZGN0EtLCQbdY7TVhZhsLDSnnDPSvrkGivQ4hl3JoHOPUlOEBIefxLKjScpH8e2+7tlRezFkk5obtIvEbC3b5X+ENqXHbsvq6RiA9E5S21ZkTUKztskPpGxPDQpKaUFPcDCSAiLpJa6XnHs4qGySh11ZTKr51oFQIDAQAB-----END RSA Public Key-----';
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