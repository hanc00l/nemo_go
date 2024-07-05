$(function () {
    //$('#btnsiderbar').click();
    load_config_server();
    load_custom('task_workspace', $('#text_task_workspace'));
    $("#buttonSaveTaskSlice").click(function () {
        if ($('#input_ipslicenumber').val() === '' || $('#input_portslicenumber').val() === '') {
            swal('Warning', "请输入数量", 'error');
            return;
        }
        $.post("/config-save-taskslice",
            {
                "portslicenumber": $('#input_portslicenumber').val(),
                "ipslicenumber": $('#input_ipslicenumber').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
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
    });
    $("#buttonSaveNotify").click(function () {
        $.post("/config-save-notify",
            {
                "token_serverchan": $('#input_serverchan').val(),
                "token_dingtalk": $('#input_dingtalk').val(),
                "token_feishu": $('#input_feishu').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
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
    });
    $("#buttonTestNotify").click(function () {
        $.post("/config-test-notify", {}, function (data, e) {
            if (e === "success" && data['status'] === 'success') {
                swal({
                    title: "测试完成",
                    text: data['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
    });
    $("#buttonSaveWikiFeishu").click(function () {
        $.post("/config-save-wikifeishu",
            {
                "feishuappid": $('#input_feishu_appid').val(),
                "feishusecret": $('#input_feishu_secret').val(),
                "feishurefreshtoken": $('#input_feishu_refreshtoken').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
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
        let encryptor = new JSEncrypt();
        encryptor.setPublicKey(pubKey);
        let oldpass = encryptor.encrypt($('#input_oldpass').val());
        let newpass = encryptor.encrypt($('#input_password1').val());
        $.post("/config-change-password",
            {
                "oldpass": oldpass,
                "newpass": newpass,
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
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
    $("#buttonRefreshToken").click(function () {
        $.post("wiki-refresh-token",
            {}, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
                    swal({
                        title: "刷新Token成功",
                        text: data['msg'],
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
    $("#buttonSaveTaskWorkspace").click(function () {
        save_custom("task_workspace", $('#text_task_workspace').val(), "/custom-save-taskworkspace")
    });
    $("#button_feishu_auth").click(function () {
        let appid = $('#input_feishu_appid').val();
        if (appid === '') {
            swal('Warning', "请输入应用ID！", 'error');
            return;
        }
        let appScret = $('#input_feishu_secret').val();
        if (appScret === '') {
            swal('Warning', "请输入应用Secret！", 'error');
            return;
        }
        let server_ip = $('#input_server_ip').val();
        if (server_ip === '') {
            swal('Warning', "请输入服务器IP！", 'error');
            return;
        }
        let server_port = $('#input_server_port').val();
        if (server_port === '') {
            swal('Warning', "请输入服务器端口！", 'error');
            return;
        }
        let callback_url = encodeURIComponent($('#select_server_protocol').val() + "://" + server_ip + ":" + server_port + "/wiki-feishu-code");
        let url = "https://open.feishu.cn/open-apis/authen/v1/index?redirect_uri=" + callback_url + "&app_id=" + appid + "&state=RANDOMSTATE";
        window.open(url, "_blank");
        $('#openfeishu').modal('hide');
    });
    $("#buttonUpdateMinichatConfig").click(function () {
        $.post("/config-update-minichat",
            {
                "notdelfiledir": $('#checkbox_notdelfiledir').is(":checked"),
                "loadhistory": $('#checkbox_loadhistory').is(":checked"),
                "maxhistorymessage": $('#input_maxhistorymessage').val(),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
                    swal({
                        title: "更新成功！",
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
});

function load_config_server() {
    $.post("/config-list-server", function (data) {
        $('#input_ipslicenumber').val(data['ipslicenumber']);
        $('#input_portslicenumber').val(data['portslicenumber']);
        $('#nemo_version').html(data['version']);

        $('#input_serverchan').val(data['serverchan']);
        $('#input_dingtalk').val(data['dingtalk']);
        $('#input_feishu').val(data['feishu']);

        $('#input_feishu_appid').val(data['feishuappid']);
        $('#input_feishu_secret').val(data['feishusecret']);
        $('#input_feishu_refreshtoken').val(data['feishurefreshtoken']);

        $('#checkbox_notdelfiledir').prop("checked", data['notdelfiledir']);
        $('#checkbox_loadhistory').prop("checked", data['loadhistory']);
        $('#input_maxhistorymessage').val(data['maxhistorymessage']);
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

function save_custom(type, content, url = "/custom-save") {
    if (type === "") {
        swal('Warning', "类型为空！", 'error');
        return;
    }
    $.post(url,
        {
            "type": type,
            "content": content
        }, function (data, e) {
            if (e === "success" && data['status'] === 'success') {
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
