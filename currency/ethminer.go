package currency

func init() {
	Currencies["Etherum2"] = &Etherum2{}
}

type Etherum2 struct {
}

func (e *Etherum2) Mine(config CurrencyConfig) {

}

/*

#if ETH_STRATUM
	void doStratum()
	{
		map<string, Farm::SealerDescriptor> sealers;
#if ETH_ETHASHCL
		sealers["opencl"] = Farm::SealerDescriptor{ &CLMiner::instances, [](FarmFace& _farm, unsigned _index){ return new CLMiner(_farm, _index); } };
#endif
#if ETH_ETHASHCUDA
		sealers["cuda"] = Farm::SealerDescriptor{ &CUDAMiner::instances, [](FarmFace& _farm, unsigned _index){ return new CUDAMiner(_farm, _index); } };
#endif
		if (!m_farmRecheckSet)
			m_farmRecheckPeriod = m_defaultStratumFarmRecheckPeriod;

		Farm f;

#if API_CORE
		Api api(this->m_api_port, f);
#endif

		// this is very ugly, but if Stratum Client V2 tunrs out to be a success, V1 will be completely removed anyway
		if (m_stratumClientVersion == 1) {
			EthStratumClient client(&f, m_minerType, m_farmURL, m_port, m_user, m_pass, m_maxFarmRetries, m_worktimeout, m_stratumProtocol, m_email);
			if (m_farmFailOverURL != "")
			{
				if (m_fuser != "")
				{
					client.setFailover(m_farmFailOverURL, m_fport, m_fuser, m_fpass);
				}
				else
				{
					client.setFailover(m_farmFailOverURL, m_fport);
				}
			}
			f.setSealers(sealers);

			f.onSolutionFound([&](Solution sol)
			{
				if (client.isConnected()) {
					client.submit(sol);
				}
				else {
					cwarn << "Can't submit solution: Not connected";
				}
				return false;
			});
			f.onMinerRestart([&](){
				client.reconnect();
			});

			while (client.isRunning())
			{
				auto mp = f.miningProgress(m_show_hwmonitors);
				if (client.isConnected())
				{
					if (client.current())
					{
						minelog << mp << f.getSolutionStats() << f.farmLaunchedFormatted();
#if ETH_DBUS
						dbusint.send(toString(mp).data());
#endif
					}
					else
					{
						minelog << "Waiting for work package...";
					}

					if (this->m_report_stratum_hashrate) {
						auto rate = mp.rate();
						client.submitHashrate(toJS(rate));
					}
				}
				this_thread::sleep_for(chrono::milliseconds(m_farmRecheckPeriod));
			}
		}
		else if (m_stratumClientVersion == 2) {
			EthStratumClientV2 client(&f, m_minerType, m_farmURL, m_port, m_user, m_pass, m_maxFarmRetries, m_worktimeout, m_stratumProtocol, m_email);
			if (m_farmFailOverURL != "")
			{
				if (m_fuser != "")
				{
					client.setFailover(m_farmFailOverURL, m_fport, m_fuser, m_fpass);
				}
				else
				{
					client.setFailover(m_farmFailOverURL, m_fport);
				}
			}
			f.setSealers(sealers);

			f.onSolutionFound([&](Solution sol)
			{
				client.submit(sol);
				return false;
			});
			f.onMinerRestart([&](){
				client.reconnect();
			});

			while (client.isRunning())
			{
				auto mp = f.miningProgress(m_show_hwmonitors);
				if (client.isConnected())
				{
					if (client.current())
					{
						minelog << mp << f.getSolutionStats();
#if ETH_DBUS
						dbusint.send(toString(mp).data());
#endif
					}
					else if (client.waitState() == MINER_WAIT_STATE_WORK)
					{
						minelog << "Waiting for work package...";
					}

					if (this->m_report_stratum_hashrate) {
						auto rate = mp.rate();
						client.submitHashrate(toJS(rate));
					}
				}
				this_thread::sleep_for(chrono::milliseconds(m_farmRecheckPeriod));
			}
		}

	}
#endif


*/
