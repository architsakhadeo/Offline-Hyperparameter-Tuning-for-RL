from importlib import import_module

def offline_agent(parent, param):
    agent = import_module("Agents."+parent)
    agent = getattr(agent, parent)
    class OfflineAgent(agent):
        def __init__(self):
            super().__init__()
            self.offline_data = param.offline_data
            return

        def offline_start(self, state, t):
            self.start(state)
            return self.offline_data[t]

        def offline_step(self, reward, state, end_of_ep, t):
            self.step(reward, state, end_of_ep=False)
            return self.offline_data[t]
    return OfflineAgent()