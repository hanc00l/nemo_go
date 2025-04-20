$(function () {
    load_custom();

    $('#select_category').change(function () {
        load_custom();
    });
    $('#button_refresh').click(function () {
        load_custom();
    });
    $('#button_save').click(function () {
        save_custom();
    });
})

function load_custom() {
    $.ajax({
        type: 'POST',
        url: '/custom-load', // Beego处理URL
        data: {category: $('#select_category').val()},
        dataType: 'json',
        success: function (response) {
            // 请求成功，解析返回的JSON数据并填充表单
            $('#text_description').val(response.description);
            $('#text_data').val(response.data);
            $('#hidden_id').val(response.id);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function save_custom() {
    $.ajax({
        type: 'POST',
        url: '/custom-save', // Beego处理URL
        data: {
            id: $('#hidden_id').val(),
            category: $('#select_category').val(),
            description: $('#text_description').val(),
            data: $('#text_data').val(),
        },
        dataType: 'json',
        success: function (data) {
            show_response_message("保存成功", data, function () {
            })
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}