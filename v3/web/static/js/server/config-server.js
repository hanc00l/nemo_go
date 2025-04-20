$(function () {
    load_server_version();
    $("#buttonChangPassword").click(function () {
        change_password();
    });
});

function change_password() {
    if ($('#input_oldpass').val() === '' || $('#input_password1').val() === '' || $('#input_password2').val() === '') {
        swal('Warning', "请输入密码！", 'error');
        return;
    }
    if ($('#input_password1').val() !== $('#input_password2').val()) {
        swal('Warning', "两次新密码不一致！", 'error');
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/config-change-password',
        data: {
            oldpass: $('#input_oldpass').val(),
            newpass: $('#input_password1').val()
        },
        dataType: 'json',
        success: function (response) {
            show_response_message("修改密码", response, function () {
            })
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function load_server_version() {
    $.ajax({
        type: 'POST',
        url: '/config-server-version',
        dataType: 'json',
        success: function (response) {
            $('#nemo_version').text(response.msg);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}