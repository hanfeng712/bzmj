package fsm

import (
	"fmt"
)

var callBackFunc func()

// 状态接口
type IFSMState interface {
	Enter()
	Do()
	Exit()
	CheckTransition()
}

// State父struct
type FSMState struct {
	id       string
	callback callBackFunc
}

// 进入状态
func (this *FSMState) Enter() {
	//
	fmt.Println("state enter")
}

//状态处理函数
func (this *FSMState) Do() {
	fmt.Println("state do")
	this.callback()
}

// 退出状态
func (this *FSMState) Exit() {
	//
	fmt.Println("state exit")
}

//添加状态回调函数
func (this *FSMState) addStateCallBack(f callBackFunc) {
	this.callback = f
}

// 状态转移检测
func (this *FSMState) CheckTransition() {
	//
}

/*******************************************************************************************/

type FSM struct {
	// 持有状态集合
	states map[string]IFSMState
	// 当前状态
	current_state IFSMState
	// 默认状态
	default_state IFSMState
}

// 初始化FSM
func (this *FSM) Init() {
	//
}

// 设置默认的State
func (this *FSM) SetDefaultState(state IFSMState) {
	//
}

// 转移状态
func (this *FSM) TransitionState() {
	//
}

// 添加状态到FSM
func (this *FSM) AddState(key string, state IFSMState) {
	//
}

//状态切换
func (this *FSM) SwitchFsmState(stateId string) {

}

//暂停FSM
func (this *FSM) PauseStateMachine() {

}

// 重置FSM
func (this *FSM) ResetStateMachine() {
	//
}
