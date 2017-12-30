#pragma once

/*
	This file is part of cpp-ethereum.
	cpp-ethereum is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	cpp-ethereum is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with cpp-ethereum.  If not, see <http://www.gnu.org/licenses/>.
*/
/** @file MinerAux.cpp
 * @author Gav Wood <i@gavwood.com>
 * @date 2014
 * CLI module for mining.
 */


#include <thread>
#include <chrono>
#include <fstream>
#include <iostream>
#include <signal.h>
#include <random>

#include <boost/algorithm/string.hpp>
#include <boost/algorithm/string/trim_all.hpp>
#include <boost/optional.hpp>

#include <libethcore/Exceptions.h>
#include <libdevcore/SHA3.h>
#include <libethcore/EthashAux.h>
#include <libethcore/Farm.h>
#if ETH_ETHASHCL
#include <libethash-cl/CLMiner.h>
#endif
#if ETH_ETHASHCUDA
#include <libethash-cuda/CUDAMiner.h>
#endif
#if ETH_STRATUM
#include <libstratum/EthStratumClient.h>
#include <libstratum/EthStratumClientV2.h>
#endif
#if ETH_DBUS
#include "DBusInt.h"
#endif
#if API_CORE
#include <libapicore/Api.h>
#endif

namespace dev { namespace eth {} }

using namespace std;
using namespace dev;
using namespace dev::eth;


class BadArgument: public Exception {};
struct MiningChannel: public LogChannel
{
	static const char* name() { return EthGreen "  m"; }
	static const int verbosity = 2;
	static const bool debug = false;
};
#define minelog clog(MiningChannel)

inline std::string toJS(unsigned long _n)
{
	std::string h = toHex(toCompactBigEndian(_n, 1));
	// remove first 0, if it is necessary;
	std::string res = h[0] != '0' ? h : h.substr(1);
	return "0x" + res;
}

class MinerCLI
{
public:
	enum class OperationMode
	{
		None,
		Benchmark,
		Simulation,
		Farm,
		Stratum
	};

	MinerCLI(OperationMode _mode = OperationMode::None): mode(_mode) {}

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

	/// Operating mode.
	OperationMode mode;

	/// Mining options
	bool m_running = true;
	MinerType m_minerType = MinerType::Mixed;
	unsigned m_openclPlatform = 0;
	unsigned m_miningThreads = UINT_MAX;
	bool m_shouldListDevices = false;
#if ETH_ETHASHCL
	unsigned m_openclDeviceCount = 0;
	unsigned m_openclDevices[16];
	unsigned m_openclThreadsPerHash = 8;
#if !ETH_ETHASHCUDA
	unsigned m_globalWorkSizeMultiplier = CLMiner::c_defaultGlobalWorkSizeMultiplier;
	unsigned m_localWorkSize = CLMiner::c_defaultLocalWorkSize;
#endif
#endif
#if ETH_ETHASHCUDA
	unsigned m_globalWorkSizeMultiplier = ethash_cuda_miner::c_defaultGridSize;
	unsigned m_localWorkSize = ethash_cuda_miner::c_defaultBlockSize;
	unsigned m_cudaDeviceCount = 0;
	unsigned m_cudaDevices[16];
	unsigned m_numStreams = ethash_cuda_miner::c_defaultNumStreams;
	unsigned m_cudaSchedule = 4; // sync
#endif
	unsigned m_dagLoadMode = 0; // parallel
	unsigned m_dagCreateDevice = 0;
	/// Benchmarking params
	unsigned m_benchmarkWarmup = 15;
	unsigned m_parallelHash    = 4;
	unsigned m_benchmarkTrial = 3;
	unsigned m_benchmarkTrials = 5;
	unsigned m_benchmarkBlock = 0;
	/// Farm params
	string m_farmURL = "http://127.0.0.1:8545";
	string m_farmFailOverURL = "";


	string m_activeFarmURL = m_farmURL;
	unsigned m_farmRetries = 0;
	unsigned m_maxFarmRetries = 3;
	unsigned m_farmRecheckPeriod = 500;
	unsigned m_defaultStratumFarmRecheckPeriod = 2000;
	bool m_farmRecheckSet = false;
	int m_worktimeout = 180;
	bool m_show_hwmonitors = false;
#if API_CORE
	int m_api_port = 0;
#endif

#if ETH_STRATUM
	bool m_report_stratum_hashrate = false;
	int m_stratumClientVersion = 1;
	int m_stratumProtocol = STRATUM_PROTOCOL_STRATUM;
	string m_user;
	string m_pass;
	string m_port;
	string m_fuser = "";
	string m_fpass = "";
	string m_email = "";
#endif
	string m_fport = "";

#if ETH_DBUS
	DBusInt dbusint;
#endif
};