# Решение задачи о переливании компота (Reinforcement Learning)

## Описание задачи
Агент обучается переливать компот между 4 стаканами (вместимость 4 ед.) так, чтобы во всех стаканах было по 2 единицы. Реализованы алгоритмы **Q-learning** и **SARSA**.

## Параметры среды
- **Состояние**: массив из 4 чисел (0-4).
- **Действие**: переливание `amount` единиц из `src` в `dst` (48 дискретных действий).
- **Награда**: -1 за шаг, +100 за достижение цели, shaping-награда за приближение.

## Результаты обучения
Обучение проводилось на 200 000 эпизодов.

### Кривые обучения
<img width="1740" height="818" alt="1_learning_curves" src="https://github.com/user-attachments/assets/fbf0e6e9-aa52-41c9-a54f-118585552e53" />


### Примеры решения 
<img width="571" height="439" alt="image" src="https://github.com/user-attachments/assets/76d1f1be-da29-4e9a-a56a-f1d645eec5f0" />
<img width="1192" height="324" alt="2_trajectory_table" src="https://github.com/user-attachments/assets/58f22d6d-b865-4cc1-b75b-e209c28fe231" />
<img width="800" height="500" alt="4_animation" src="https://github.com/user-attachments/assets/52f5600e-c778-4098-bbfa-f9e0348820fc" />



## Сравнение методов
В данной задаче Q-learning показал более быструю сходимость по сравнению с SARSA, так как он напрямую оптимизирует оптимальную политику (off-policy), в то время как SARSA учитывает исследовательские шаги (on-policy).
