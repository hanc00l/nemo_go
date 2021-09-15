$(function () {
    $('#btnsiderbar').click();
    load_config();
    load_custom('honeypot', $('#text_honeypot'));
    load_custom('service', $('#text_service'));
    load_custom('iplocation', $('#text_iplocation'));
    load_custom('iplocationB', $('#text_iplocationB'));
    load_custom('iplocationC', $('#text_iplocationC'));

    $("#buttonSaveNmap").click(function () {
        swal('Warning', "暂不支持保存配置，请手工修改配置文件!", 'error');
        return;
    });
    $("#buttonSaveTaskSlice").click(function () {
        if ($('#input_ipslicenumber').val() === '' || $('#input_portslicenumber').val() === '') {
            swal('Warning', "请输入数量", 'error');
            return;
        }
        $.post("/config-save-taskslice",
            {
                "portslicenumber": $('#input_portslicenumber').val(),
                "ipslicenumber": $('#input_ipslicenumber').val()
            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "保存成功！",
                        text: "",
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    })
});
$("#buttonSaveHoneypot").click(function () {
    save_custom("honeypot", $('#text_honeypot').val())
});
$("#buttonSaveService").click(function () {
    save_custom("service", $('#text_service').val())
});
$("#buttonSaveIPLocation").click(function () {
    save_custom("iplocation", $('#text_iplocation').val())
});
$("#buttonSaveIPLocationB").click(function () {
    save_custom("iplocationB", $('#text_iplocationB').val())
});
$("#buttonSaveIPLocationC").click(function () {
    save_custom("iplocationC", $('#text_iplocationC').val())
});
$("#buttonChangPassword").click(function () {
    if ($('#input_oldpass').val() === '' || $('#input_password1').val() === '' || $('#input_password2').val() === '') {
        swal('Warning', "请输入密码！", 'error');
        return;
    }
    if ($('#input_password1').val() !== $('#input_password2').val()) {
        swal('Warning', "两次新密码不一致！", 'error');
        return;
    }
    $.post("/config-change-password",
        {
            "oldpass": $('#input_oldpass').val(),
            "newpass": $('#input_password1').val()
        }, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                $('#input_oldpass').val('');
                $('#input_password1').val('');
                $('#input_password2').val('');
                swal({
                    title: "密码修改成功！",
                    text: "",
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                    timer: 3000
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
});

function load_config() {
    $.post("/config-list", function (data) {
        $('#input_cmdbin').val(data['cmdbin']);
        $('#input_port').val(data['port']);
        $('#select_tech').val(data['tech']);
        $('#input_rate').val(data['rate']);
        $('#checkbox_ping').prop("checked", data['ping']);
        $('#input_ipslicenumber').val(data['ipslicenumber'])
        $('#input_portslicenumber').val(data['portslicenumber'])
    });
}

function load_custom(type, textCtl) {
    $.post("/custom-load",
        {
            "type": type,
        }, function (data, e) {
            if (e === "success" && data['status'] === "success") {
                textCtl.val(data['msg']);
            } else {
                textCtl.val('load custom fail!');
            }
        });
}

function save_custom(type, content) {
    if (content === "") {
        swal('Warning', "空的配置！", 'error');
        return;
    }
    $.post("/custom-save",
        {
            "type": type,
            "content": content
        }, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                swal({
                    title: "保存成功！",
                    text: "",
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                    timer: 3000
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
}