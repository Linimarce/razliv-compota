package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	NumGlasses = 4
	MaxLevel   = 4
	GoalLevel  = 2
	MaxSteps   = 100
	NumActions = NumGlasses * (NumGlasses - 1) * MaxLevel
)

type State [NumGlasses]int

func (s State) IsTerminal() bool {
	for _, v := range s {
		if v != GoalLevel {
			return false
		}
	}
	return true
}

func (s State) String() string {
	return fmt.Sprintf("[%d %d %d %d]", s[0], s[1], s[2], s[3])
}

func DistanceToGoal(s State) int {
	dist := 0
	for _, v := range s {
		d := v - GoalLevel
		if d < 0 {
			d = -d
		}
		dist += d
	}
	return dist
}

// --- Кодирование действий (src, dst, amount) ---
func encodeAction(src, dst, amount int) int {
	if dst > src {
		dst--
	}
	return (src*(NumGlasses-1)+dst)*MaxLevel + (amount - 1)
}

func decodeAction(act int) (int, int, int) {
	amount := act%MaxLevel + 1
	act /= MaxLevel
	src := act / (NumGlasses - 1)
	dst := act % (NumGlasses - 1)
	if dst >= src {
		dst++
	}
	return src, dst, amount
}

// --- Среда ---
type Environment struct {
	noise float64
}

func NewEnvironment(noise float64) *Environment {
	return &Environment{noise: noise}
}

func (env *Environment) Reset() State {
	return RandomReachableState()
}

func (env *Environment) Step(s State, src, dst, amount int) (State, float64, bool) {
	if src == dst || s[src] == 0 || s[dst] == MaxLevel {
		return s, -5.0, false
	}
	maxTransfer := int(math.Min(float64(s[src]), float64(MaxLevel-s[dst])))
	if amount <= 0 || amount > maxTransfer {
		return s, -5.0, false
	}

	newS := s
	if rand.Float64() < env.noise {
		amount--
		if amount == 0 {
			return s, -5.0, false
		}
	}
	newS[src] -= amount
	newS[dst] += amount

	if newS == s {
		return s, -5.0, false
	}

	reward := -1.0
	oldDist := DistanceToGoal(s)
	newDist := DistanceToGoal(newS)
	if newDist < oldDist {
		reward += 0.1 * float64(oldDist-newDist)
	} else if newDist > oldDist {
		reward -= 0.1 * float64(newDist-oldDist)
	}

	done := newS.IsTerminal()
	if done {
		reward += 100.0
	}
	return newS, reward, done
}

func RandomState() State {
	var s State
	remaining := NumGlasses * GoalLevel
	for i := 0; i < NumGlasses-1; i++ {
		maxVal := int(math.Min(float64(MaxLevel), float64(remaining)))
		val := rand.Intn(maxVal + 1)
		s[i] = val
		remaining -= val
	}
	s[NumGlasses-1] = remaining
	rand.Shuffle(NumGlasses, func(i, j int) { s[i], s[j] = s[j], s[i] })
	return s
}

func IsReachable(start State) bool {
	if start.IsTerminal() {
		return true
	}
	env := NewEnvironment(0.0)
	queue := []State{start}
	visited := map[State]bool{start: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for src := 0; src < NumGlasses; src++ {
			for dst := 0; dst < NumGlasses; dst++ {
				if src == dst || cur[src] == 0 || cur[dst] == MaxLevel {
					continue
				}
				maxTransfer := int(math.Min(float64(cur[src]), float64(MaxLevel-cur[dst])))
				for amount := 1; amount <= maxTransfer; amount++ {
					next, _, _ := env.Step(cur, src, dst, amount)
					if next == cur || visited[next] {
						continue
					}
					if next.IsTerminal() {
						return true
					}
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}
	}
	return false
}

var reachableStates []State

func InitReachableStates() {
	for a := 0; a <= MaxLevel; a++ {
		for b := 0; b <= MaxLevel; b++ {
			for c := 0; c <= MaxLevel; c++ {
				for d := 0; d <= MaxLevel; d++ {
					if a+b+c+d != NumGlasses*GoalLevel {
						continue
					}
					s := State{a, b, c, d}
					if s.IsTerminal() {
						continue
					}
					if IsReachable(s) {
						reachableStates = append(reachableStates, s)
					}
				}
			}
		}
	}
	fmt.Println("Найдено решаемых стартовых состояний:", len(reachableStates))
}

func RandomReachableState() State {
	if len(reachableStates) == 0 {
		return State{1, 2, 1, 4}
	}
	return reachableStates[rand.Intn(len(reachableStates))]
}

// --- Агент ---
type Agent struct {
	Q map[State][NumActions]float64
}

func NewAgent() *Agent {
	return &Agent{Q: make(map[State][NumActions]float64)}
}

func (a *Agent) GetQ(s State, act int) float64 {
	if _, ok := a.Q[s]; !ok {
		a.Q[s] = [NumActions]float64{}
	}
	return a.Q[s][act]
}

func (a *Agent) SetQ(s State, act int, val float64) {
	if _, ok := a.Q[s]; !ok {
		a.Q[s] = [NumActions]float64{}
	}
	arr := a.Q[s]
	arr[act] = val
	a.Q[s] = arr
}

func (a *Agent) ChooseAction(s State, epsilon float64) int {
	validActions := []int{}
	for src := 0; src < NumGlasses; src++ {
		for dst := 0; dst < NumGlasses; dst++ {
			if src == dst || s[src] == 0 || s[dst] == MaxLevel {
				continue
			}
			maxTransfer := int(math.Min(float64(s[src]), float64(MaxLevel-s[dst])))
			for amount := 1; amount <= maxTransfer; amount++ {
				validActions = append(validActions, encodeAction(src, dst, amount))
			}
		}
	}
	if len(validActions) == 0 {
		return 0
	}
	if rand.Float64() < epsilon {
		return validActions[rand.Intn(len(validActions))]
	}
	bestAct := validActions[0]
	bestVal := -math.MaxFloat64
	for _, act := range validActions {
		v := a.GetQ(s, act)
		if v > bestVal {
			bestVal = v
			bestAct = act
		}
	}
	return bestAct
}

func epsilonAt(ep, episodes int, epsilon0, epsilonMin float64) float64 {
	frac := float64(ep) / float64(episodes)
	return math.Max(epsilonMin, epsilon0-(epsilon0-epsilonMin)*frac)
}

// --- Q-Learning ---
func TrainQLearning(env *Environment, episodes int, alpha0, gamma, epsilon0, epsilonMin float64) ([]float64, *Agent) {
	agent := NewAgent()
	rewards := make([]float64, episodes)
	for ep := 0; ep < episodes; ep++ {
		s := env.Reset()
		totalReward := 0.0
		epsilon := epsilonAt(ep, episodes, epsilon0, epsilonMin)
		for step := 0; step < MaxSteps; step++ {
			if s.IsTerminal() {
				break
			}
			act := agent.ChooseAction(s, epsilon)
			src, dst, amount := decodeAction(act)
			nextS, reward, done := env.Step(s, src, dst, amount)

			maxNextQ := 0.0
			if !done {
				maxNextQ = -math.MaxFloat64
				for a := 0; a < NumActions; a++ {
					v := agent.GetQ(nextS, a)
					if v > maxNextQ {
						maxNextQ = v
					}
				}
			}
			currentQ := agent.GetQ(s, act)
			agent.SetQ(s, act, currentQ+alpha0*(reward+gamma*maxNextQ-currentQ))
			totalReward += reward
			s = nextS
			if done {
				break
			}
		}
		rewards[ep] = totalReward
	}
	return rewards, agent
}

// --- SARSA ---
func TrainSARSA(env *Environment, episodes int, alpha0, gamma, epsilon0, epsilonMin float64) ([]float64, *Agent) {
	agent := NewAgent()
	rewards := make([]float64, episodes)
	for ep := 0; ep < episodes; ep++ {
		s := env.Reset()
		totalReward := 0.0
		epsilon := epsilonAt(ep, episodes, epsilon0, epsilonMin)
		act := agent.ChooseAction(s, epsilon)
		for step := 0; step < MaxSteps; step++ {
			if s.IsTerminal() {
				break
			}
			src, dst, amount := decodeAction(act)
			nextS, reward, done := env.Step(s, src, dst, amount)
			var nextAct int
			var nextQ float64
			if done {
				nextQ = 0
			} else {
				nextAct = agent.ChooseAction(nextS, epsilon)
				nextQ = agent.GetQ(nextS, nextAct)
			}
			currentQ := agent.GetQ(s, act)
			agent.SetQ(s, act, currentQ+alpha0*(reward+gamma*nextQ-currentQ))
			totalReward += reward
			s = nextS
			if done {
				break
			}
			act = nextAct
		}
		rewards[ep] = totalReward
	}
	return rewards, agent
}

// --- Демонстрация с сохранением траектории и состояний ---
func ShowPolicy(agent *Agent, env *Environment) {
	fmt.Println("\n=== Демонстрация выученной политики ===")
	s := env.Reset()
	fmt.Println("Старт:", s)

	states := [][4]int{ {s[0], s[1], s[2], s[3]} }
	var trajectory [][4]int // step, src, dst, amount

	steps := 0
	for !s.IsTerminal() && steps < 30 {
		bestAct := -1
		bestVal := -math.MaxFloat64
		for src := 0; src < NumGlasses; src++ {
			for dst := 0; dst < NumGlasses; dst++ {
				if src == dst || s[src] == 0 || s[dst] == MaxLevel {
					continue
				}
				maxTransfer := int(math.Min(float64(s[src]), float64(MaxLevel-s[dst])))
				for amount := 1; amount <= maxTransfer; amount++ {
					act := encodeAction(src, dst, amount)
					v := agent.GetQ(s, act)
					if v > bestVal {
						bestVal = v
						bestAct = act
					}
				}
			}
		}
		if bestAct == -1 {
			fmt.Println("Нет допустимых действий.")
			break
		}

		src, dst, amount := decodeAction(bestAct)
		fmt.Printf("DEBUG: выбрано src=%d dst=%d amount=%d Q=%.2f\n", src, dst, amount, bestVal)

		newS := s
		newS[src] -= amount
		newS[dst] += amount

		if newS == s {
			fmt.Println("DEBUG: действие не меняет состояние, ищем fallback")
			bestDist := DistanceToGoal(s)
			found := false
			for src2 := 0; src2 < NumGlasses; src2++ {
				for dst2 := 0; dst2 < NumGlasses; dst2++ {
					if src2 == dst2 || s[src2] == 0 || s[dst2] == MaxLevel {
						continue
					}
					maxTransfer := int(math.Min(float64(s[src2]), float64(MaxLevel-s[dst2])))
					for amount2 := 1; amount2 <= maxTransfer; amount2++ {
						testS := s
						testS[src2] -= amount2
						testS[dst2] += amount2
						d := DistanceToGoal(testS)
						if d < bestDist {
							bestDist = d
							src, dst, amount = src2, dst2, amount2
							found = true
						}
					}
				}
			}
			if !found {
				fmt.Println("Нет действия, уменьшающего расстояние. Тупик.")
				break
			}
			newS = s
			newS[src] -= amount
			newS[dst] += amount
		}

		fmt.Printf("Шаг %d: переливаем %d из %d в %d (Q=%.2f) -> ", steps, amount, src, dst, bestVal)
		s = newS
		fmt.Println(s)

		trajectory = append(trajectory, [4]int{steps, src, dst, amount})
		states = append(states, [4]int{s[0], s[1], s[2], s[3]})
		steps++
		if s.IsTerminal() {
			fmt.Println("ЦЕЛЬ ДОСТИГНУТА!")
		}
	}
	if !s.IsTerminal() {
		fmt.Println("Цель не достигнута за 30 шагов.")
	}

	// --- Сохранение trajectory.csv ---
	tf, _ := os.Create("/content/trajectory.csv")
	defer tf.Close()
	tf.WriteString("step,src,dst,amount\n")
	for _, t := range trajectory {
		tf.WriteString(fmt.Sprintf("%d,%d,%d,%d\n", t[0], t[1], t[2], t[3]))
	}

	// --- Сохранение states.csv ---
	sf, _ := os.Create("/content/states.csv")
	defer sf.Close()
	sf.WriteString("step,g0,g1,g2,g3\n")
	for i, st := range states {
		sf.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n", i, st[0], st[1], st[2], st[3]))
	}

	fmt.Println("Сохранено: /content/trajectory.csv, /content/states.csv")
}

// --- Сохранение Q-таблицы в JSON ---
func SaveQTable(agent *Agent, path string) {
	data := map[string]map[string]float64{}
	for s, arr := range agent.Q {
		key := fmt.Sprintf("%d,%d,%d,%d", s[0], s[1], s[2], s[3])
		data[key] = map[string]float64{}
		for a, v := range arr {
			if v != 0 {
				data[key][fmt.Sprintf("%d", a)] = v
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("Ошибка сохранения Q-таблицы:", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(data)
	fmt.Println("Q-таблица сохранена:", path)
}

// --- BFS ---
func BFSPath(start State) [][3]int {
	type node struct {
		s    State
		path [][3]int
	}
	queue := []node{{start, nil}}
	visited := map[State]bool{start: true}
	env := NewEnvironment(0.0)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.s.IsTerminal() {
			return cur.path
		}
		for src := 0; src < NumGlasses; src++ {
			for dst := 0; dst < NumGlasses; dst++ {
				if src == dst || cur.s[src] == 0 || cur.s[dst] == MaxLevel {
					continue
				}
				maxTransfer := int(math.Min(float64(cur.s[src]), float64(MaxLevel-cur.s[dst])))
				for amount := 1; amount <= maxTransfer; amount++ {
					next, _, _ := env.Step(cur.s, src, dst, amount)
					if visited[next] {
						continue
					}
					visited[next] = true
					newPath := append(append([][3]int{}, cur.path...), [3]int{src, dst, amount})
					queue = append(queue, node{next, newPath})
				}
			}
		}
	}
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	InitReachableStates()

	episodes := 200000
	alpha := 0.2
	gamma := 0.95
	epsilon0 := 1.0
	epsilonMin := 0.4

	fmt.Println("Training Q-Learning...")
	envQ := NewEnvironment(0.0)
	qRewards, qAgent := TrainQLearning(envQ, episodes, alpha, gamma, epsilon0, epsilonMin)

	fmt.Println("Training SARSA...")
	envS := NewEnvironment(0.0)
	sRewards, _ := TrainSARSA(envS, episodes, alpha, gamma, epsilon0, epsilonMin)

	// rewards.csv
	f, err := os.Create("/content/rewards.csv")
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer f.Close()
	f.WriteString("Episode,QLearning,SARSA\n")
	for i := 0; i < episodes; i++ {
		f.WriteString(fmt.Sprintf("%d,%f,%f\n", i, qRewards[i], sRewards[i]))
	}
	fmt.Println("Done! Results saved to /content/rewards.csv")

	// Сохраняем Q-таблицу
	SaveQTable(qAgent, "/content/qtable.json")

	// Демонстрация + сохранение trajectory.csv и states.csv
	envDemo := NewEnvironment(0.0)
	ShowPolicy(qAgent, envDemo)

	// BFS для сравнения
	fmt.Println("\n=== BFS: кратчайший путь ===")
	start := envDemo.Reset()
	fmt.Println("Старт:", start)
	path := BFSPath(start)
	if path == nil {
		fmt.Println("BFS: путь не найден.")
	} else {
		s := start
		for i, step := range path {
			fmt.Printf("Шаг %d: переливаем %d из %d в %d -> ", i, step[2], step[0], step[1])
			s, _, _ = envDemo.Step(s, step[0], step[1], step[2])
			fmt.Println(s)
		}
		fmt.Println("ЦЕЛЬ ДОСТИГНУТА (BFS)!")
	}
}