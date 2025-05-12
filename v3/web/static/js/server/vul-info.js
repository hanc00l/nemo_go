$(function () {
    load_vul_data();
})

function fill_form_with_data(data) {
    $('#vul_pocfile').text(data.pocfile);
    $('#vul_name').text(data.name);
    $('#vul_url').text(data.url);
    $('#vul_authority').text(data.authority);
    $('#vul_severity').text(data.severity);
    $('#vul_source').text(data.source);
    $('#vul_extra').text(data.extra);
    $('#vul_create_time').text(data.create_time);
    $('#vul_update_time').text(data.update_time);
}

function load_vul_data() {
    const urlParams = new URLSearchParams(window.location.search);
    const id = urlParams.get('id');

    if (!isNotEmpty(id)) {
        alert("ID未提供");
        return;
    }
    $.ajax({
        type: 'POST', url: '/vul-info', // Beego处理URL
        data: {id: id},
        dataType: 'json',
        success: function (response) {
            fill_form_with_data(response);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}
