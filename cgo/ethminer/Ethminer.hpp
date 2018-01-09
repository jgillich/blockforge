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
#include <libstratum/EthStratumClientV2.h>

using namespace std;
using namespace dev;
using namespace dev::eth;

class Ethminer
{
public:
  Ethminer(string url, string port, string user, string pass, string email, vector<unsigned> openclDevices, vector<unsigned> cudaDevices)
  {
    // log errors only
    g_logVerbosity = 0;

    MinerType minerType;

    if (openclDevices.size() > 0)
    {
#if ETH_ETHASHCL
      CLMiner::setDevices(&openclDevices[0], openclDevices.size());

      CLMiner::setThreadsPerHash(m_openclThreadsPerHash);
      if (!CLMiner::configureGPU(
              CLMiner::c_defaultLocalWorkSize,
              CLMiner::c_defaultGlobalWorkSizeMultiplier,
              m_openclPlatform,
              0,
              0,
              0))
        exit(1);
      CLMiner::setNumInstances(openclDevices.size());

      minerType = MinerType::CL;

#else
      cerr << "Selected GPU mining without having compiled with -DETHASHCL=1" << endl;
      exit(1);
#endif
    }

    if (cudaDevices.size() > 0)
    {
#if ETH_ETHASHCUDA
      CUDAMiner::setDevices(&cudaDevices[0], cudaDevices.size());

      CUDAMiner::setNumInstances(cudaDevices.size());
      if (!CUDAMiner::configureGPU(
              ethash_cuda_miner::c_defaultBlockSize,
              ethash_cuda_miner::c_defaultGridSize,
              ethash_cuda_miner::c_defaultNumStreams,
              4,
              0,
              0,
              0))
        exit(1);

      CUDAMiner::setParallelHash(m_parallelHash);

      if (minerType == MinerType::CL)
      {
        minerType = MinerType::Mixed;
      }
      else
      {
        minerType = MinerType::CUDA;
      }
#else
      cerr << "CUDA support disabled. Configure project build with -DETHASHCUDA=ON" << endl;
      exit(1);
#endif
    }

    map<string, Farm::SealerDescriptor> sealers;
#if ETH_ETHASHCL
    sealers["opencl"] = Farm::SealerDescriptor{&CLMiner::instances, [](FarmFace &_farm, unsigned _index) { return new CLMiner(_farm, _index); }};
#endif
#if ETH_ETHASHCUDA
    sealers["cuda"] = Farm::SealerDescriptor{&CUDAMiner::instances, [](FarmFace &_farm, unsigned _index) { return new CUDAMiner(_farm, _index); }};
#endif

    Farm f;

    EthStratumClientV2 client(&f, minerType, url, port, user, pass, m_maxFarmRetries, m_worktimeout, STRATUM_PROTOCOL_STRATUM, email);
    f.setSealers(sealers);

    f.onSolutionFound([&](Solution sol) {
      client.submit(sol);
      return false;
    });
    f.onMinerRestart([&]() {
      client.reconnect();
    });

    while (client.isRunning())
    {

      auto mp = f.miningProgress(true);

      if (client.isConnected())
      {
        m_hashrate = mp.rate();
        /*
        int gpuIndex = 0;
        int numGpus = mp.minersHashes.size();
        for (auto const &i : p.minersHashes)
        {
          detailedMhEth << std::fixed << std::setprecision(0) << (p.minerRate(i) / 1000.0f) << (((numGpus - 1) > gpuIndex) ? ";" : "");
          detailedMhDcr << "off" << (((numGpus - 1) > gpuIndex) ? ";" : ""); // DualMining not supported
          gpuIndex++;
        }

        gpuIndex = 0;
        numGpus = p.minerMonitors.size();
        for (auto const &i : p.minerMonitors)
        {
          tempAndFans << i.tempC << ";" << i.fanP << (((numGpus - 1) > gpuIndex) ? "; " : ""); // Fetching Temp and Fans
          gpuIndex++;
        }
        */
      }

      this_thread::sleep_for(chrono::milliseconds(m_farmRecheckPeriod));
    }
  }

  int hashrate()
  {
    return m_hashrate;
  }

private:
  //vector<uint64_t> m_hashrate;
   int m_hashrate;


  unsigned m_openclPlatform = 0;
  unsigned m_openclThreadsPerHash = 8;
  unsigned m_maxFarmRetries = 3;
  unsigned m_farmRecheckPeriod = 2000;
  bool m_farmRecheckSet = false;
  int m_worktimeout = 180;
};
