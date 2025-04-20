$(function () {
    load_tree_data();
});

function load_tree_data() {
    const urlParams = new URLSearchParams(window.location.search);
    const maintaskId = urlParams.get('maintaskId'); // 假设URL参数名为id

    if (!isNotEmpty(maintaskId)) {
        alert("任务ID未提供");
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/maintask-tree-data',
        data: {maintaskId: maintaskId},
        dataType: 'json',
        success: function (data) {
            init_tree(data);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function defaultNodeTemplate2(content) {
    return `<div style='display: flex; justify-content: center; align-items: center; text-align: center; height: 100%; word-wrap: break-word; overflow-wrap: break-word; white-space: normal; width: 100%;' title='${content}'>${content}</div>`;
}

function defaultNodeTemplate(content) {
    return `<div style='display: flex;justify-content: center;align-items: center; text-align: center; height: 100%;' title='${content}'>${content}</div>`;
}

function init_tree(data) {
    const options = {
        height: 700,
        nodeWidth: 250,
        nodeHeight: 80,
        childrenSpacing: 100,
        siblingSpacing: 30,
        direction: 'top',
        nodeTemplate: defaultNodeTemplate,
        canvasStyle: 'border: 1px solid black; background: #f6f6f6;',
    };

    const tree = new ApexTree(document.getElementById('svg-tree'), options);
    const graph = tree.render(data);
    document.getElementById('layoutTop').addEventListener('click', (e) => {
        graph.changeLayout('top');
    });
    document.getElementById('layoutBottom').addEventListener('click', (e) => {
        graph.changeLayout('bottom');
    });

    document.getElementById('layoutLeft').addEventListener('click', (e) => {
        graph.changeLayout('left');
    });
    document.getElementById('layoutRight').addEventListener('click', (e) => {
        graph.changeLayout('right');
    });

    document.getElementById('fitScreen').addEventListener('click', (e) => {
        graph.fitScreen();
    });
}