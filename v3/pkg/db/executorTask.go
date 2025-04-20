package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type ExecutorTask struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type ExecuteTaskDocument struct {
	Id          bson.ObjectID `bson:"_id"`
	WorkspaceId string        `bson:"workspaceId" json:"workspaceId"`

	TaskId        string  `bson:"taskId" json:"taskId"`
	MainTaskId    string  `bson:"mainTaskId" json:"mainTaskId"`
	Executor      string  `bson:"executor,omitempty" json:"executor,omitempty"`
	Target        string  `bson:"target" json:"target"`
	ExcludeTarget string  `bson:"excludeTarget,omitempty" json:"excludeTarget,omitempty"`
	Args          string  `bson:"args" json:"args"`
	Status        string  `bson:"status" json:"status"`
	Worker        string  `bson:"worker,omitempty" json:"worker,omitempty"`
	Result        string  `bson:"result,omitempty" json:"result,omitempty"`
	Progress      string  `bson:"progress,omitempty" json:"progress,omitempty"`
	ProgressRate  float64 `bson:"progressRate,omitempty" json:"progressRate,omitempty"`
	PreTaskId     string  `bson:"preTaskId" json:"preTaskId"`

	StartTime  *time.Time `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime    *time.Time `bson:"end_time,omitempty"  json:"end_time,omitempty"`
	CreateTime time.Time  `bson:"create_time" json:"create_time"`
	UpdateTime time.Time  `bson:"update_time" json:"update_time"`
}
type AggregationResult struct {
	Executor string        `bson:"_id"`
	Statuses []StatusCount `bson:"statuses"`
}

type StatusCount struct {
	Status string `bson:"status"`
	Count  int    `bson:"count"`
}

// ApexTreeNode 定义ApexTree需要的节点结构
type ApexTreeNode struct {
	Id       string          `json:"id"`
	Name     string          `json:"name"`
	Children []*ApexTreeNode `json:"children,omitempty"`
}

// TaskChainResult 包含任务链查询的所有结果
type TaskChainResult struct {
	RootTasks []string        // 根任务ID列表(可能有多个)
	ApexTree  []*ApexTreeNode // ApexTree节点数组
}

func NewExecutorTask(client *mongo.Client) *ExecutorTask {
	return &ExecutorTask{
		DatabaseName:   GlobalDatabase,
		CollectionName: "executorTask",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (t *ExecutorTask) Insert(doc ExecuteTaskDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	// 插入文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	_, err = col.InsertOne(t.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (t *ExecutorTask) GetById(id string) (result ExecuteTaskDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	idd, _ := bson.ObjectIDFromHex(id)
	filter := bson.M{ID: idd}
	err = col.FindOne(t.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	return
}

func (t *ExecutorTask) GetByTaskId(taskId string) (result ExecuteTaskDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	filter := bson.M{TaskId: taskId}
	err = col.FindOne(t.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	return
}

func (t *ExecutorTask) Find(filter bson.M, page, pageSize int) (result []ExecuteTaskDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	// 计算分页
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64(page * pageSize))
	}
	opts.SetSort(bson.M{UpdateTime: -1})
	cur, err := col.Find(t.Ctx, filter, opts)
	// 查询
	if err != nil {
		return
	}
	defer cur.Close(t.Ctx)

	if err = cur.All(t.Ctx, &result); err != nil {
		return nil, err
	}
	return
}

func (t *ExecutorTask) Update(id string, update bson.M) (isSuccess bool, err error) {
	// 更新文档
	idd, _ := bson.ObjectIDFromHex(id)
	filter := bson.M{ID: idd}
	update[UpdateTime] = time.Now()
	updateDoc := bson.M{"$set": update}
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	result, err := col.UpdateOne(t.Ctx, filter, updateDoc)
	if err != nil {
		return false, err
	}
	if result.ModifiedCount > 0 {
		isSuccess = true
	}
	return
}

func (t *ExecutorTask) Delete(id string) (isSuccess bool, err error) {
	_idd, _ := bson.ObjectIDFromHex(id)
	// 删除文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	result, err := col.DeleteOne(t.Ctx, bson.M{ID: _idd})
	if err != nil {
		return false, err
	}
	if result.DeletedCount > 0 {
		isSuccess = true
	}
	return
}

func (t *ExecutorTask) DeleteByMainTaskId(maintaskId string) (isSuccess bool, err error) {
	// 删除文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	result, err := col.DeleteMany(t.Ctx, bson.M{MainTaskId: maintaskId})
	if err != nil {
		return false, err
	}
	if result.DeletedCount > 0 {
		isSuccess = true
	}
	return
}

func (t *ExecutorTask) AggregateMainTaskProgress(mainTaskId string) (result []AggregationResult, err error) {
	collection := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	pipeline := mongo.Pipeline{
		bson.D{
			{"$match", bson.M{
				"mainTaskId": mainTaskId,
			}},
		},
		bson.D{
			{"$group", bson.D{
				{"_id", bson.D{
					{"executor", "$executor"},
					{"status", "$status"},
				}},
				{"count", bson.M{"$sum": 1}},
			}},
		},
		bson.D{
			{"$group", bson.D{
				{"_id", "$_id.executor"},
				{"statuses", bson.M{"$push": bson.D{
					{"status", "$_id.status"},
					{"count", "$count"},
				}}},
			}},
		},
		bson.D{
			{"$sort", bson.M{"_id": 1}},
		},
	}

	cur, err := collection.Aggregate(t.Ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(t.Ctx)

	if err = cur.All(t.Ctx, &result); err != nil {
		return nil, err
	}
	return
}

// GetDailyStats 获取最近N天的统计信息
func (t *ExecutorTask) GetDailyStats(workspaceId string, days int) (ApexChartData, error) {
	collection := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)

	// 1. 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1) // 包含今天

	// 2. 执行聚合查询
	pipeline := mongo.Pipeline{
		// 匹配指定日期范围内的记录
		{{"$match", bson.D{
			{"workspaceId", workspaceId},
			{"update_time", bson.D{
				{"$gte", startDate},
				{"$lte", endDate},
			}},
		}}},
		// 按天分组统计
		{{"$group", bson.D{
			{"_id", bson.D{
				{"day", bson.D{
					{"$dateToString", bson.D{
						{"format", "%Y-%m-%d"},
						{"date", "$update_time"},
					}},
				}},
			}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		// 按日期排序
		{{"$sort", bson.D{
			{"_id.day", 1},
		}}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return ApexChartData{}, fmt.Errorf("aggregation failed: %v", err)
	}
	defer cursor.Close(ctx)

	// 3. 处理聚合结果
	var dailyStats []DailyStat
	if err := cursor.All(ctx, &dailyStats); err != nil {
		return ApexChartData{}, fmt.Errorf("cursor decode failed: %v", err)
	}

	// 4. 生成完整日期范围并填充数据
	return generateChartData(startDate, endDate, dailyStats), nil
}

// GetTaskChainByMainTaskID 主函数，根据mainTaskId获取所有子任务信息
func (t *ExecutorTask) GetTaskChainByMainTaskID(mainTaskID string) (*TaskChainResult, error) {
	collection := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)

	// 1. 获取所有子任务并按创建时间排序
	tasks, err := getSubTasksByMainTaskID(t.Ctx, mainTaskID, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub tasks: %v", err)
	}

	// 2. 构建任务执行链路
	childMap, rootTasks := buildTaskChain(tasks)

	// 3. 生成ApexTree结构
	apexTree, err := generateApexTree(tasks, childMap, rootTasks, mainTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate apex tree: %v", err)
	}

	return &TaskChainResult{
		RootTasks: rootTasks,
		ApexTree:  apexTree,
	}, nil
}

// getSubTasksByMainTaskID 获取所有子任务并按创建时间排序
func getSubTasksByMainTaskID(ctx context.Context, mainTaskID string, collection *mongo.Collection) ([]ExecuteTaskDocument, error) {
	filter := bson.M{"mainTaskId": mainTaskID}
	opts := options.Find().SetSort(bson.D{{Key: "create_time", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %v", err)
	}
	defer cursor.Close(ctx)

	var tasks []ExecuteTaskDocument
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %v", err)
	}

	return tasks, nil
}

// buildTaskChain 构建任务执行链路
func buildTaskChain(tasks []ExecuteTaskDocument) (map[string][]string, []string) {
	// 找出所有根任务(没有preTaskID或preTaskId为空的任务)
	var rootTasks []string
	childMap := make(map[string][]string) // preTaskID -> []taskIDs

	for _, task := range tasks {
		if task.PreTaskId == "" {
			rootTasks = append(rootTasks, task.TaskId)
		} else {
			childMap[task.PreTaskId] = append(childMap[task.PreTaskId], task.TaskId)
		}
	}

	return childMap, rootTasks
}

// buildApexTree 递归构建ApexTree结构
func buildApexTree(nodeID string, taskMap map[string]ExecuteTaskDocument, childMap map[string][]string) *ApexTreeNode {
	task := taskMap[nodeID]
	node := &ApexTreeNode{
		Id:   task.TaskId,
		Name: fmt.Sprintf("%s %s", task.Executor, task.Target),
	}
	// 递归添加子节点
	if children, ok := childMap[nodeID]; ok {
		for _, childID := range children {
			childNode := buildApexTree(childID, taskMap, childMap)
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

// generateApexTree 生成ApexTree结构
func generateApexTree(tasks []ExecuteTaskDocument, childMap map[string][]string, rootTasks []string, mainTaskID string) ([]*ApexTreeNode, error) {
	// 创建任务ID到任务的映射
	taskMap := make(map[string]ExecuteTaskDocument)
	for _, task := range tasks {
		taskMap[task.TaskId] = task
	}

	// 创建根节点
	rootNode := &ApexTreeNode{
		Id:   mainTaskID,
		Name: "任务开始",
	}

	// 将所有顶层任务作为根节点的子节点
	for _, rootID := range rootTasks {
		childNode := buildApexTree(rootID, taskMap, childMap)
		rootNode.Children = append(rootNode.Children, childNode)
	}

	return []*ApexTreeNode{rootNode}, nil
}
