$(function () {
    $('#btnsiderbar').click();
    // 点击关闭大图显示
    $('.imgPreview').click(function () {
        $('.imgPreview').hide();
    });
});

/**
 * 删除操作
 */
function delete_op(url, id) {
    swal({
            title: "确定要删除?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: url + '?id=' + id,
                success: function (data) {
                    location.reload();
                },
                error: function (xhr, type) {
                }
            });
        });
}

function delete_port_attr(id) {
    delete_op('/port-attr-delete', id);
}

function delete_domain_attr(id) {
    delete_op('/domain-attr-delete', id);
}

function delete_domain_fofa_attr(id) {
    delete_op('/domain-fofa-attr-delete', id);
}


function html2Escape(sHtml) {
    var temp = document.createElement("div");
    (temp.textContent != null) ? (temp.textContent = sHtml) : (temp.innerText = sHtml);
    var output = temp.innerHTML.replace(/\"/g, "&quot;").replace(/\'/g, "&acute;");
    ;
    ;
    temp = null;
    //output = output

    return output;
}

//标记颜色
function _mark_color_tag(obj_type, r_id, color) {
    var url = "/" + obj_type + "-color-tag?r_id=" + r_id;
    $.post(url,
        {
            "color": color
        }, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                location.reload();
            } else {
                swal('Warning', "标记IP颜色失败!", 'error');
            }
        });
}

//编辑备忘信息
function _edit_memo_content(obj_type) {
    const r_id = $('#r_id').val();
    var url = "/" + obj_type + "-memo-get?r_id=" + r_id;
    $.get(url,
        function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                $('#text_content').val(data['msg']);
            }
        })
    $('#editMemo').modal('toggle');
}

//保存备忘信息
function _save_memo_content(obj_type) {
    const r_id = $('#r_id').val();
    const memo = $('#text_content').val();
    var url = "/" + obj_type + "-memo-update?r_id=" + r_id;
    $.post(url,
        {
            "memo": memo
        }, function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                $('#memo_content').html("<pre>" + html2Escape(memo) + "</pre>");
                $('#editMemo').modal('hide');
            } else {
                swal('Warning', "保存失败!", 'error');
            }
        });
}

function mark_ip_color_tag(r_id, color) {
    _mark_color_tag('ip', r_id, color);
}

function mark_domain_color_tag(r_id, color) {
    _mark_color_tag('domain', r_id, color);
}

$("#btn_editIpMemo").click(function () {
    _edit_memo_content('ip');
});
$("#btn_editDomainMemo").click(function () {
    _edit_memo_content('domain');
});
$("#btn_saveIpMemo").click(function () {
    _save_memo_content('ip');
});
$("#btn_saveDomainMemo").click(function () {
    _save_memo_content('domain');
});

function show_bigpic(src) {
    $('.imgPreview img').attr('src', src);
    $('.imgPreview').show();
}

function refresh_info(type, r_id, status) {
    let url = type + "-info?" + type + "=" + r_id + "&&disable_fofa=";
    if(status){
        url += "false"
    }else{
        url += "true"
    }
    window.location.href = url;
}