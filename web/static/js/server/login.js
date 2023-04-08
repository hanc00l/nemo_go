const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyK7ChnOvz0TwoPYwzNgdkK/hVX2YtjDNr/1ODS35Q7OuINhiWn8qmr12Ex/V/K6Mm+cDnZAx98Y81PngORZ4EfOF16+7fidyflMN99oM3VNtK3amAVXDCtH7NgP/1IEk2b2yR/lqA+Lgr4qfxhBNjjCusEYsSkMr69M4/u/J+5dfKWxFN9ZQzwdGIwAkvHM40bhxZos4Yb1EsBcFm8HVtFXvJcDe72+RggqKjp3p4PMndeYDNht1dxCG5Sfa9dvbdbdzxjgMPMVt2IZaihZU+t9/0H3bvTqdc/liY3Ky6WZOhlimDpBVEiOSBFzOjwbvOt9asyZjL6bb9OMEzttO9QIDAQAB-----END RSA Public Key-----';
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