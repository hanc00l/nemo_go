const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtchoNXqWgrgoPHA/LJPLBQ8D4tHa2impOrQAfb1gGzhPJ6AquL0sGY2XJt9xw9C7Ez8CnbSNxoG5uaztyG35BgixcIgiND3ctjyRbFh+24E9guHa19roC0GjpQW86wd7EmgfAjr9yzeKhryYfoSTgUEkpNwl4d3JGL9WQbhFgfxpc35rfUK3bzCKGEvwUOHFXHC2wSSbAxcXApLPnQpvSI6NDyKuFSsOPmPSrb5rrPg4iHaOeatpEXM+NCYeNzMEzPWxsMzEK6FFzak9SJvvnEn8OU4GMDjXQURcxj3ghP189LgKoj0kgbA+z0LmYiv6Kpa3RuLseQHfaMaDU4DRNwIDAQAB-----END RSA Public Key-----';
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