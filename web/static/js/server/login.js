const pubKey = '-----BEGIN RSA Public Key-----MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAumsXgPVWrmcs6p4SLtWpusAq+MS/eZPMTPNumXNroAOnbhj9sQGVTGy4jX2INfQ+ZBsWoR8SELKwxeBNRZ0EAB2nPsweus3evz2FYfGhmLgcFPtM31Fd4igTbnGxf/rRrJE44qsa1vYb/rgU5GMlao6uFRkaO2vwG51H+0cfj7u0bxg8IcIXoTaZkQYMulqqGU+C4DaJs2f8YDFSLSbG6WvgNAg72hYxO/GMBT5SDCboHaBIMJ2768HNlWE3yPfiHG5IppA7gSCpaM6NATf7Aop/0S1tF4ZdFkq7/gArqdagpAs7Emnlc9CSS5mgjVYHCWCQGq0fNxjZDrnpoXMQjQIDAQAB-----END RSA Public Key-----';
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