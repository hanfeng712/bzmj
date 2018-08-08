package fsm

import (
	"container/list"
	"fmt"
	"reflect"
)

type callBackFunc func(arg interface{})

// State父struct
type FSMState struct {
	id       string                    //状态对应的键值
	callback callBackFunc              //普通的回调函数
	funName  string                    //对象对调方法的名字
	object   interface{}               //创建状态机的对象
	methods  map[string]reflect.Method //创装状态机对象的所有方法的集合
	arg      interface{}               //回调方法的参数
}

/*
创建状态机的状态方法
@param1 id:      状态对应的id
@param2: fun:    普通函数回调函数,兼容接口
@param3: arg:    回调函数的参数
@param4: object: 创建状态机的对象
@param5: name:   对象的回调方法，兼容接口
注意：两个兼容的回调函数，区别是：一个是函数，一个是创建状态机的对象的方法
*/
func CreateMatchineState(id string, fun callBackFunc, arg interface{}, object interface{}, name string) *FSMState {

	methods := make(map[string]reflect.Method)
	if object != nil {
		typ := reflect.TypeOf(object)
		for m := 0; m < typ.NumMethod(); m++ {
			method := typ.Method(m)
			mname := method.Name
			methods[mname] = method
		}
	}

	return &FSMState{
		id:       id,
		callback: fun,
		arg:      arg,
		object:   object,
		methods:  methods,
		funName:  name,
	}
}

/*
状态方法：进入状态
*/
func (this *FSMState) Enter() {
	//
	fmt.Println("state enter")
}

/*
状态方法：状态处理函数
*/
func (this *FSMState) Do() {
	fmt.Printf("FSMState do - key : %s\n", this.id)
	if this.object != nil {

		params := make([]reflect.Value, 2)
		params[0] = reflect.ValueOf(this.object)
		params[1] = reflect.ValueOf(this.arg)
		this.methods[this.funName].Func.Call(params)
		return
	}
	this.callback(this.arg)
}

/*
状态方法：退出状态
*/
func (this *FSMState) Exit() {
	//
	fmt.Println("state exit")
}

/*
状态方法：添加状态回调函数
*/
func (this *FSMState) addStateCallBack(f callBackFunc) {
	this.callback = f
}

/*
状态方法：状态转移检测
*/
func (this *FSMState) CheckTransition() {
	//
}

/*
设置回调参数值
*/
func (this *FSMState) setCallBackArg(arg interface{}) {

	this.arg = arg
}

/*******************************************************************************************/
func CreateFSM() *FSM {
	it := &FSM{}
	it.Init()
	return it
}

type FSM struct {
	// 持有状态集合
	statesMap map[string]*FSMState
	//
	action *list.List
	// 当前状态
	current_state *FSMState
	// 下一个状态
	next_state *FSMState
	// 默认状态
	default_state *FSMState
	runState      int
}

/*
状态机方法：初始化FSM
*/
func (this *FSM) Init() {
	//
	this.runState = 0
	this.statesMap = make(map[string]*FSMState)
	this.action = list.New()
}

/*
状态机方法：启动状态机
*/
func (this *FSM) Start() {
	fmt.Printf("FSM-Start():enter")
	if this.action.Len() == 0 {
		fmt.Printf("FSM-Start():leave1")
		return
	}
	firstKey := this.action.Front()
	this.current_state = this.statesMap[firstKey.Value.(string)]
	var index string = this.CalcNextStateKey(this.current_state)
	this.next_state = this.statesMap[index]

	this.runState = 1
	this.DoFsmState()
	fmt.Printf("FSM-Start():leave2")
}

/*
设状态机方法：置默认的State
@param: state:状态
*/
func (this *FSM) SetDefaultState(state *FSMState) {
	this.default_state = state
}

/*
添状态机方法：加状态到FSM
@param1 : 状态对应的key
@param : state:状态
*/
func (this *FSM) AddState(key string, state *FSMState) {
	this.statesMap[key] = state
	this.action.PushBack(key)
}

/*
根状态机方法：据key获取状态
@param: key:状态的Id
*/
func (this *FSM) GetStateById(key string) *FSMState {
	return this.statesMap[key]
}

/*
获状态机方法：取当前状态
*/
func (this *FSM) GetCurrentState() *FSMState {
	return this.current_state
}

/*
获状态机方法：取下一下状态
*/
func (this *FSM) GetNextState() *FSMState {
	return this.next_state
}

/*
状状态机方法：态切换
*/
func (this *FSM) DoFsmState() {
	if this.runState != 1 {
		return
	}
	//执行当前状态的动作
	this.current_state.Do()
}

func (this *FSM) SwitchFsmState() {
	this.current_state = this.next_state
	var index string = this.CalcNextStateKey(this.current_state)
	this.next_state = this.statesMap[index]
	this.DoFsmState()
}

/*
状态机方法：计算下一个状态
@parma1: crrrent:当前的状态
*/
func (this *FSM) CalcNextStateKey(current *FSMState) string {
	var index string = this.default_state.id
	for e := this.action.Front(); e != nil; e = e.Next() {
		if e.Value == current.id {
			if e.Next() != nil {
				index = (e.Next().Value).(string)
			}
			break
		}
	}
	return index
}

/*
状态机方法：设置状态的执行参数
@param1: key:状态对应的Id
@param2: arg:状态的回调函数执行参数
*/
func (this *FSM) SetStateCallBackArg(key string, arg interface{}) {
	state := this.GetStateById(key)
	if state == nil {
		return
	}
	state.setCallBackArg(arg)
}

/*
暂状态机方法：停FSM
*/
func (this *FSM) PauseStateMachine() {
	this.runState = 2
}

/*
重状态机方法：置FSM
*/
func (this *FSM) ResetStateMachine() {
	this.runState = 1
	this.current_state = this.default_state
	// 下一个状态
	var index string = this.CalcNextStateKey(this.current_state)
	this.next_state = this.statesMap[index]
}
