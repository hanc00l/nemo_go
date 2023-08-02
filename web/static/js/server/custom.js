$(function () {
    $('#btnsiderbar').click();
    load_custom('honeypot', $('#text_honeypot'));
    load_custom('service', $('#text_service'));
    load_custom('iplocation', $('#text_iplocation'));
    $('#select_iplocation_filename').change(function () {
        load_custom($('#select_iplocation_filename').val(), $('#text_iplocation'));
    });
    load_custom('config.xray', $('#text_xray_config'));
    $('#select_xray_config_filename').change(function () {
        load_custom($('#select_xray_config_filename').val(), $('#text_xray_config'));
    });
    load_custom('black_domain', $('#text_black_domain_ip'));
    $('#select_black_filename').change(function () {
        load_custom($('#select_black_filename').val(), $('#text_black_domain_ip'));
    });
    load_custom('task_workspace', $('#text_task_workspace'));
    load_custom('fofa_filter_keyword', $('#text_fofa_filter_keyword'));
    $('#select_fofa_filter_keyword_type').change(function () {
        load_custom($('#select_fofa_filter_keyword_type').val(), $('#text_fofa_filter_keyword'));
    });
    $("#buttonSaveHoneypot").click(function () {
        save_custom("honeypot", $('#text_honeypot').val())
    });
    $("#buttonSaveService").click(function () {
        save_custom("service", $('#text_service').val())
    });
    $("#buttonSaveIPLocation").click(function () {
        save_custom($("#select_iplocation_filename").val(), $('#text_iplocation').val())
    });
    $("#buttonSaveBlackDomainIP").click(function () {
        save_custom($("#select_black_filename").val(), $('#text_black_domain_ip').val())
    });
    $("#buttonSaveTaskWorkspace").click(function () {
        save_custom("task_workspace", $('#text_task_workspace').val(), "/custom-save-taskworkspace")
    });
    $("#buttonSaveFOFAFilterKeyword").click(function () {
        save_custom($("#select_fofa_filter_keyword_type").val(), $('#text_fofa_filter_keyword').val())
    });
    $("#buttonUploadPoc").click(function () {
        let formData = new FormData();
        formData.append('type', $('#select_poc_type').val());
        formData.append('file', $('#file_poc')[0].files[0]);
        if (formData.get("file") === "undefined") {
            swal('Warning', '请选择要上传的文件！', 'error');
            return;
        }
        $.ajax({
            url: "/config-upload-poc",
            type: "post",
            data: formData,
            //十分重要，不能省略
            cache: false,
            processData: false,
            contentType: false,
            success: function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "保存成功！",
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
            }
        });
    });
    $("#buttonSaveXrayConfig").click(function () {
        save_custom($('#select_xray_config_filename').val(), $('#text_xray_config').val())
    });
});

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
